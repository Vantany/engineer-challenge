package email

import (
	"log"
)

type Logger interface {
	Printf(format string, v ...any)
}

type LoggingEmailService struct {
	logger Logger
}

func NewLoggingEmailService(logger Logger) *LoggingEmailService {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingEmailService{logger: logger}
}

func (s *LoggingEmailService) SendResetToken(email, token string) error {
	s.logger.Printf("send password reset token: email=%s token=%s", email, token)
	return nil
}

