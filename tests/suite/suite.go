package suite

import (
	"context"
	ssov1 "github.com/jasmanchik/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"
)

type Suit struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const grpsHost = "localhost"

func New(t *testing.T) (context.Context, *Suit) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local_test.yaml")

	ctx, ctxCancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	t.Cleanup(func() {
		t.Helper()
		ctxCancel()
	})

	cc, err := grpc.DialContext(context.Background(), grpcAddress(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc connection failed: %v", err)
	}

	return ctx, &Suit{Cfg: cfg, T: t, AuthClient: ssov1.NewAuthClient(cc)}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpsHost, strconv.Itoa(cfg.GRPC.Port))
}
