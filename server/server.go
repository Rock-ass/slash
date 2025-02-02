package server

import (
	"context"
	"fmt"
	"net"
	"time"

	apiv1 "github.com/boojack/slash/api/v1"
	apiv2 "github.com/boojack/slash/api/v2"
	"github.com/boojack/slash/server/profile"
	"github.com/boojack/slash/store"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	e *echo.Echo

	Profile *profile.Profile
	Store   *store.Store

	// API services.
	apiV2Service *apiv2.APIV2Service
}

func NewServer(ctx context.Context, profile *profile.Profile, store *store.Store) (*Server, error) {
	e := echo.New()
	e.Debug = true
	e.HideBanner = true
	e.HidePort = true

	s := &Server{
		e:       e,
		Profile: profile,
		Store:   store,
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}",` +
			`"method":"${method}","uri":"${uri}",` +
			`"status":${status},"error":"${error}"}` + "\n",
	}))

	e.Use(middleware.Gzip())

	e.Use(middleware.CORS())

	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Request timeout",
		Timeout:      30 * time.Second,
	}))

	embedFrontend(e)

	// In dev mode, we'd like to set the const secret key to make signin session persistence.
	secret := "slash"
	if profile.Mode == "prod" {
		var err error
		secret, err = s.getSystemSecretSessionName(ctx)
		if err != nil {
			return nil, err
		}
	}

	rootGroup := e.Group("")
	// Register API v1 routes.
	apiV1Service := apiv1.NewAPIV1Service(profile, store)
	apiV1Service.Start(rootGroup, secret)

	s.apiV2Service = apiv2.NewAPIV2Service(secret, profile, store, s.Profile.Port+1)
	// Register gRPC gateway as api v2.
	if err := s.apiV2Service.RegisterGateway(ctx, e); err != nil {
		return nil, fmt.Errorf("failed to register gRPC gateway: %w", err)
	}

	// Register resource service.
	resourceService := NewResourceService(profile, store)
	resourceService.Register(rootGroup)

	return s, nil
}

func (s *Server) Start(_ context.Context) error {
	// Start gRPC server.
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Profile.Port+1))
	if err != nil {
		return err
	}
	go func() {
		if err := s.apiV2Service.GetGRPCServer().Serve(listen); err != nil {
			println("grpc server listen error")
		}
	}()

	return s.e.Start(fmt.Sprintf(":%d", s.Profile.Port))
}

func (s *Server) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Shutdown echo server.
	if err := s.e.Shutdown(ctx); err != nil {
		fmt.Printf("failed to shutdown server, error: %v\n", err)
	}

	// Close database connection.
	if err := s.Store.Close(); err != nil {
		fmt.Printf("failed to close database, error: %v\n", err)
	}

	fmt.Printf("server stopped properly\n")
}

func (s *Server) GetEcho() *echo.Echo {
	return s.e
}

func (s *Server) getSystemSecretSessionName(ctx context.Context) (string, error) {
	secretSessionNameValue, err := s.Store.GetWorkspaceSetting(ctx, &store.FindWorkspaceSetting{
		Key: store.WorkspaceDisallowSignUp,
	})
	if err != nil {
		return "", err
	}
	if secretSessionNameValue == nil || secretSessionNameValue.Value == "" {
		tempSecret := uuid.New().String()
		secretSessionNameValue, err = s.Store.UpsertWorkspaceSetting(ctx, &store.WorkspaceSetting{
			Key:   store.WorkspaceSecretSessionName,
			Value: string(tempSecret),
		})
		if err != nil {
			return "", err
		}
	}
	return secretSessionNameValue.Value, nil
}
