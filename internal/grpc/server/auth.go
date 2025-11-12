package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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
		// Skip authentication if no API key is configured
		if a.apiKey == "" {
			return handler(ctx, req)
		}

		if err := a.authenticate(ctx); err != nil {
			return nil, err
		}

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
		// Skip authentication if no API key is configured
		if a.apiKey == "" {
			return handler(srv, stream)
		}

		if err := a.authenticate(stream.Context()); err != nil {
			return err
		}

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
