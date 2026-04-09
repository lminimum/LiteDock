package grpc

import (
	v1 "github.com/evrone/go-clean-template/internal/controller/grpc/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewRouter -.
func NewRouter(app *pbgrpc.Server, t usecase.Translation, auth usecase.Auth, l logger.Interface) {
	{
		v1.NewTranslationRoutes(app, t, l)
	}

	{
		v1.NewAuthRoutes(app, auth, l)
	}

	reflection.Register(app)
}
