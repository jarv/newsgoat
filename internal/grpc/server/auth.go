package server

import (
	"context"
	"log/slog"
	"os"
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var logger *slog.Logger

func init() {
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	slogHandler := slog.NewJSONHandler(os.Stdout, slogOpts)
	logger = slog.New(slogHandler)
}

const apiKeyHeader = "x-api-key"

// AuthInterceptor implements API key authentication for gRPC
type AuthInterceptor struct {
	apiKey string
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(apiKey string) *AuthInterceptor {
	return &AuthInterceptor{
		apiKey: apiKey,
	}
}

// Unary returns a server interceptor function for unary RPCs
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract method name and client info for logging
		method := path.Base(info.FullMethod)
		clientAddr := getClientAddr(ctx)

		// Skip authentication if no API key is configured
		if a.apiKey == "" {
			logger.Info("Command received", "method", method, "client", clientAddr)
			return handler(ctx, req)
		}

		if err := a.authenticate(ctx); err != nil {
			logger.Warn("Authentication failed", "method", method, "client", clientAddr, "error", err.Error())
			return nil, err
		}

		logger.Info("Command received", "method", method, "client", clientAddr)
		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function for streaming RPCs
func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract method name and client info for logging
		method := path.Base(info.FullMethod)
		ctx := stream.Context()
		clientAddr := getClientAddr(ctx)

		// Skip authentication if no API key is configured
		if a.apiKey == "" {
			logger.Info("Stream started", "method", method, "client", clientAddr)
			return handler(srv, stream)
		}

		if err := a.authenticate(ctx); err != nil {
			logger.Warn("Authentication failed", "method", method, "client", clientAddr, "error", err.Error())
			return err
		}

		logger.Info("Stream started", "method", method, "client", clientAddr)
		return handler(srv, stream)
	}
}

func (a *AuthInterceptor) authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(apiKeyHeader)
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing API key")
	}

	if values[0] != a.apiKey {
		return status.Errorf(codes.Unauthenticated, "invalid API key")
	}

	return nil
}

// getClientAddr extracts the client address from the context
func getClientAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "unknown"
}
