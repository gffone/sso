package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"time"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
