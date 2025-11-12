package server

import (
	"context"
	"database/sql"

	"github.com/jarv/newsgoat/internal/database"
	pb "github.com/jarv/newsgoat/internal/grpc/gen/newsgoat/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SettingsService implements the gRPC SettingsService
type SettingsService struct {
	pb.UnimplementedSettingsServiceServer
	queries *database.Queries
}

// NewSettingsService creates a new SettingsService
func NewSettingsService(queries *database.Queries) *SettingsService {
	return &SettingsService{
		queries: queries,
	}
}

func dbSettingToPb(setting database.Setting) *pb.Setting {
	return &pb.Setting{
		Key:       setting.Key,
		Value:     setting.Value,
		UpdatedAt: timestamppb.New(setting.UpdatedAt.Time),
	}
}

func (s *SettingsService) GetSetting(ctx context.Context, req *pb.GetSettingRequest) (*pb.GetSettingResponse, error) {
	setting, err := s.queries.GetSetting(ctx, req.Key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "setting not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get setting: %v", err)
	}

	return &pb.GetSettingResponse{
		Setting: dbSettingToPb(setting),
	}, nil
}

func (s *SettingsService) SetSetting(ctx context.Context, req *pb.SetSettingRequest) (*pb.SetSettingResponse, error) {
	err := s.queries.SetSetting(ctx, database.SetSettingParams{
		Key:   req.Key,
		Value: req.Value,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set setting: %v", err)
	}

	return &pb.SetSettingResponse{}, nil
}

func (s *SettingsService) GetAllSettings(ctx context.Context, req *pb.GetAllSettingsRequest) (*pb.GetAllSettingsResponse, error) {
	settings, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get settings: %v", err)
	}

	pbSettings := make([]*pb.Setting, len(settings))
	for i, setting := range settings {
		pbSettings[i] = dbSettingToPb(setting)
	}

	return &pb.GetAllSettingsResponse{
		Settings: pbSettings,
	}, nil
}

func (s *SettingsService) DeleteSetting(ctx context.Context, req *pb.DeleteSettingRequest) (*pb.DeleteSettingResponse, error) {
	err := s.queries.DeleteSetting(ctx, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete setting: %v", err)
	}

	return &pb.DeleteSettingResponse{}, nil
}
