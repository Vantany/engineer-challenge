package grpc

import (
	"context"
	"log/slog"
	"time"

	"auth-service/internal/transport/ratelimit"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func RateLimitInterceptor(limiter *ratelimit.RouteLimiter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to extract peer info")
		}

		ip := p.Addr.String()

		if !limiter.Allow(ip, info.FullMethod) {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}

		return handler(ctx, req)
	}
}

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		var remoteAddr string
		if p, ok := peer.FromContext(ctx); ok {
			remoteAddr = p.Addr.String()
		}

		statusCode := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				statusCode = s.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		span := trace.SpanFromContext(ctx)
		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		// Предотвращаем двойное логирование для запросов через grpc-gateway
		// grpc-gateway добавляет специфичные заголовки
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if agents := md.Get("grpcgateway-user-agent"); len(agents) > 0 {
				// Глушим лог так как HTTP middleware уже залогировал его
				return resp, err
			}
			// grpc-gateway также устанавливает специфичный user-agent
			if fwds := md.Get("x-forwarded-host"); len(fwds) > 0 {
				return resp, err
			}
		}

		attrs := []any{
			"method", info.FullMethod,
			"remote_addr", remoteAddr,
			"duration", time.Since(start),
			"status", statusCode,
			"trace_id", traceID,
		}

		if err != nil {
			if statusCode == codes.Internal || statusCode == codes.Unknown {
				attrs = append(attrs, "error_details", err.Error())
				slog.Error("gRPC Request Failed", attrs...)
			} else {
				attrs = append(attrs, "error", err.Error())
				slog.Warn("gRPC Request Client Error", attrs...)
			}
		} else {
			slog.Info("gRPC Request", attrs...)
		}

		return resp, err
	}
}
