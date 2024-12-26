package suit

import (
	"context"
	ssov1 "github.com/gffone/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"
)

const (
	grpcHost = "localhost"
)

type Suit struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

func NewSuit(t *testing.T) (context.Context, *Suit) {
	t.Helper()
	t.Parallel()

	ctx := context.Background()
	cfg := config.MustLoad()
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.NewClient(grpcAddr(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	return ctx, &Suit{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}

}

func grpcAddr(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
