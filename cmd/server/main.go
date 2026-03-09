package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/db"
	"auth-service/internal/infrastructure/email"
	"auth-service/internal/infrastructure/jwt"
	"auth-service/internal/infrastructure/postgres"
	"auth-service/internal/infrastructure/tracing"
	"auth-service/internal/logger"
	transportgrpc "auth-service/internal/transport/grpc"
	authv1 "auth-service/internal/transport/grpc/gen/api/auth/v1"
	transporthttp "auth-service/internal/transport/http"
	"auth-service/internal/transport/ratelimit"
	"auth-service/internal/usecase/commands"
	"auth-service/internal/usecase/queries"
	"auth-service/internal/worker"

	"strings"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func main() {
	logg := logger.Setup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	tp, err := tracing.InitTracer(context.Background(), "auth-service", cfg.OTLPEndpoint)
	if err != nil {
		logg.Error("failed to init tracer", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logg.Error("failed to shutdown tracer", "error", err)
			}
		}()
	}

	dbCfg := db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	sqlxDB, err := db.NewPostgres(dbCfg)
	if err != nil {
		logg.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer sqlxDB.Close()

	userRepo := postgres.NewUserRepository(sqlxDB)
	sessionRepo := postgres.NewSessionRepository(sqlxDB)
	resetTokenRepo := postgres.NewResetTokenRepository(sqlxDB)
	outboxRepo := postgres.NewOutboxRepository(sqlxDB)
	txManager := postgres.NewTxManager(sqlxDB)

	tokenService, err := jwt.NewTokenService(jwt.Config{
		PrivateKeyPath: cfg.PrivateKey,
		PublicKeyPath:  cfg.PublicKey,
	})
	if err != nil {
		logg.Error("failed to init token service", "error", err)
		os.Exit(1)
	}

	passwordPolicy := domain.NewDefaultPasswordPolicy()

	stdLogger := slog.NewLogLogger(logg.Handler(), slog.LevelInfo)
	emailService := email.NewLoggingEmailService(stdLogger)

	now := func() time.Time { return time.Now().UTC() }

	outboxWorker := worker.NewOutboxWorker(outboxRepo, emailService, logg, 5*time.Second)
	go outboxWorker.Run(ctx)

	registerHandler := &commands.RegisterHandler{
		Users:          userRepo,
		Sessions:       sessionRepo,
		TxManager:      txManager,
		PasswordPolicy: passwordPolicy,
		TokenService:   tokenService,
		Now:            now,
	}

	loginHandler := &commands.LoginHandler{
		Users:          userRepo,
		Sessions:       sessionRepo,
		TxManager:      txManager,
		PasswordPolicy: passwordPolicy,
		TokenService:   tokenService,
		Now:            now,
	}

	logoutHandler := &commands.LogoutHandler{
		Sessions:     sessionRepo,
		TokenService: tokenService,
		Now:          now,
	}

	refreshHandler := &commands.RefreshTokenHandler{
		Users:        userRepo,
		Sessions:     sessionRepo,
		TokenService: tokenService,
		Now:          now,
	}

	requestResetHandler := &commands.RequestPasswordResetHandler{
		Users:        userRepo,
		ResetTokens:  resetTokenRepo,
		Outbox:       outboxRepo,
		TxManager:    txManager,
		TokenService: tokenService,
		Now:          now,
	}

	resetPasswordHandler := &commands.ResetPasswordHandler{
		Users:          userRepo,
		ResetTokens:    resetTokenRepo,
		Sessions:       sessionRepo,
		TxManager:      txManager,
		PasswordPolicy: passwordPolicy,
		TokenService:   tokenService,
		Now:            now,
	}

	validateTokenHandler := &queries.ValidateTokenHandler{
		Tokens: tokenService,
		Users:  userRepo,
	}

	limiter := ratelimit.NewRouteLimiter(100, time.Minute)
	limiter.AddRoute("/auth.v1.AuthService/Login", 5, time.Minute)
	limiter.AddRoute("/api/v1/auth/login", 5, time.Minute)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcprometheus.UnaryServerInterceptor,
			transportgrpc.LoggingInterceptor(),
			transportgrpc.RateLimitInterceptor(limiter),
		),
	)

	grpcprometheus.Register(grpcServer)

	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%s", cfg.MetricsPort)
		logg.Info("Metrics server listening", "addr", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil && err != http.ErrServerClosed {
			logg.Error("Metrics server failed", "error", err)
		}
	}()

	authHandler := &transportgrpc.AuthHandler{
		RegisterHandler:      registerHandler,
		LoginHandler:         loginHandler,
		LogoutHandler:        logoutHandler,
		RefreshHandler:       refreshHandler,
		RequestResetHandler:  requestResetHandler,
		ResetPasswordHandler: resetPasswordHandler,
		ValidateTokenHandler: validateTokenHandler,
	}

	authv1.RegisterAuthServiceServer(grpcServer, authHandler)

	reflection.Register(grpcServer)

	grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logg.Error("failed to listen for grpc", "error", err)
		os.Exit(1)
	}

	go func() {
		logg.Info("gRPC server listening", "addr", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logg.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.ToLower(key) == "authorization" {
				return "authorization", true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			reqID := r.Header.Get("X-Request-Id")
			traceID := trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()

			s, _ := status.Convert(err).WithDetails(&errdetails.ErrorInfo{
				Reason: "request_context",
				Domain: "auth-service",
				Metadata: map[string]string{
					"request_id": reqID,
					"trace_id":   traceID,
				},
			})

			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, s.Err())
		}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, gwMux, grpcAddr, opts)
	if err != nil {
		logg.Error("failed to register gateway", "error", err)
		os.Exit(1)
	}

	httpMux := http.NewServeMux()

	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	httpMux.Handle("/", gwMux)

	handler := transporthttp.CORSMiddleware(
		transporthttp.LoggingMiddleware(
			transporthttp.RateLimitMiddleware(limiter)(httpMux),
		),
	)

	handler = otelhttp.NewHandler(handler, "http-gateway")

	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logg.Info("HTTP Gateway listening", "addr", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("http gateway failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()

	grpcServer.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logg.Error("error shutting down http gateway", "error", err)
	}
}
