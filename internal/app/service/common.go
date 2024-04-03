package service

import (
	"github.com/aleks0ps/gophermart/internal/app/storage"
	"go.uber.org/zap"
)

type Service struct {
	Logger     *zap.SugaredLogger
	DB         storage.Storage
	AccrualURL string
}

func (s *Service) AccrualHTTP() string {
	return "http://" + s.AccrualURL
}
