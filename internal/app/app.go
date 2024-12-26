package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
	"time"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, tokenTTL time.Duration) *App {
	storage, err := postgres.New()
	if err != nil {
		panic(err)
	}

	authService := auth.NewAuth(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
