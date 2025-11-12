package server

import (
	"context"
	"database/sql"

	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/feeds"
	pb "github.com/jarv/newsgoat/internal/grpc/gen/newsgoat/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FeedService implements the gRPC FeedService
type FeedService struct {
	pb.UnimplementedFeedServiceServer
	manager *feeds.Manager
}

// NewFeedService creates a new FeedService
func NewFeedService(manager *feeds.Manager) *FeedService {
	return &FeedService{
		manager: manager,
	}
}

// Helper functions to convert between database types and protobuf types

func dbFeedToPb(feed database.Feed) *pb.Feed {
	pbFeed := &pb.Feed{
		Id:          feed.ID,
		Url:         feed.Url,
		Title:       feed.Title,
		Description: feed.Description,
		Visible:     feed.Visible,
	}

	if feed.LastUpdated.Valid {
		pbFeed.LastUpdated = timestamppb.New(feed.LastUpdated.Time)
	}
	if feed.LastError.Valid {
		pbFeed.LastError = feed.LastError.String
	}
	if feed.LastErrorTime.Valid {
		pbFeed.LastErrorTime = timestamppb.New(feed.LastErrorTime.Time)
	}
	if feed.CreatedAt.Valid {
		pbFeed.CreatedAt = timestamppb.New(feed.CreatedAt.Time)
	}
	if feed.Etag.Valid {
		pbFeed.Etag = feed.Etag.String
	}
	if feed.LastModified.Valid {
		pbFeed.LastModified = feed.LastModified.String
	}
	if feed.CacheControlMaxAge.Valid {
		pbFeed.CacheControlMaxAge = feed.CacheControlMaxAge.Int64
	}

	return pbFeed
}

func dbItemWithReadStatusToPb(item database.GetItemsWithReadStatusRow) *pb.ItemWithReadStatus {
	pbItem := &pb.Item{
		Id:          item.ID,
		FeedId:      item.FeedID,
		Guid:        item.Guid,
		Title:       item.Title,
		Description: item.Description,
		Content:     item.Content,
		Link:        item.Link,
	}

	if item.Published.Valid {
		pbItem.Published = timestamppb.New(item.Published.Time)
	}
	if item.CreatedAt.Valid {
		pbItem.CreatedAt = timestamppb.New(item.CreatedAt.Time)
	}

	return &pb.ItemWithReadStatus{
		Item: pbItem,
		Read: item.Read,
	}
}

func dbFeedStatsToPb(stats database.GetFeedStatsRow) *pb.FeedStats {
	pbStats := &pb.FeedStats{
		Id:          stats.ID,
		Title:       stats.Title,
		Url:         stats.Url,
		TotalItems:  stats.TotalItems,
		UnreadItems: stats.UnreadItems,
	}

	if stats.LastError.Valid {
		pbStats.LastError = stats.LastError.String
	}
	if stats.LastErrorTime.Valid {
		pbStats.LastErrorTime = timestamppb.New(stats.LastErrorTime.Time)
	}

	return pbStats
}

func dbLogMessageToPb(msg database.LogMessage) *pb.LogMessage {
	pbMsg := &pb.LogMessage{
		Id:      msg.ID,
		Level:   msg.Level,
		Message: msg.Message,
	}

	if msg.Timestamp.Valid {
		pbMsg.Timestamp = timestamppb.New(msg.Timestamp.Time)
	}
	if msg.Attributes.Valid {
		pbMsg.Details = msg.Attributes.String
	}

	return pbMsg
}

// High-level feed operations

func (s *FeedService) AddFeed(ctx context.Context, req *pb.AddFeedRequest) (*pb.AddFeedResponse, error) {
	if err := s.manager.AddFeed(req.Url); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add feed: %v", err)
	}

	// Retrieve the feed we just added to return it
	feed, err := s.manager.GetFeedByURL(req.Url)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "feed added but failed to retrieve: %v", err)
	}

	return &pb.AddFeedResponse{
		Feed: dbFeedToPb(feed),
	}, nil
}

func (s *FeedService) AddFeedWithoutFetching(ctx context.Context, req *pb.AddFeedWithoutFetchingRequest) (*pb.AddFeedWithoutFetchingResponse, error) {
	if err := s.manager.AddFeedWithoutFetching(req.Url); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add feed: %v", err)
	}

	// Retrieve the feed we just added to return it
	feed, err := s.manager.GetFeedByURL(req.Url)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "feed added but failed to retrieve: %v", err)
	}

	return &pb.AddFeedWithoutFetchingResponse{
		Feed: dbFeedToPb(feed),
	}, nil
}

func (s *FeedService) RefreshFeed(ctx context.Context, req *pb.RefreshFeedRequest) (*pb.RefreshFeedResponse, error) {
	if err := s.manager.RefreshFeed(req.FeedId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refresh feed: %v", err)
	}

	return &pb.RefreshFeedResponse{}, nil
}

func (s *FeedService) RefreshFeedByURL(ctx context.Context, req *pb.RefreshFeedByURLRequest) (*pb.RefreshFeedByURLResponse, error) {
	if err := s.manager.RefreshFeedByURL(req.Url); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refresh feed: %v", err)
	}

	return &pb.RefreshFeedByURLResponse{}, nil
}

func (s *FeedService) RefreshAllFeeds(ctx context.Context, req *pb.RefreshAllFeedsRequest) (*pb.RefreshAllFeedsResponse, error) {
	if err := s.manager.RefreshAllFeeds(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refresh feeds: %v", err)
	}

	return &pb.RefreshAllFeedsResponse{}, nil
}

// Feed CRUD operations

func (s *FeedService) CreateFeed(ctx context.Context, req *pb.CreateFeedRequest) (*pb.CreateFeedResponse, error) {
	// This is a low-level operation - not typically used. Use AddFeed instead.
	return nil, status.Error(codes.Unimplemented, "use AddFeed instead")
}

func (s *FeedService) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.GetFeedResponse, error) {
	feed, err := s.manager.GetFeed(req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "feed not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get feed: %v", err)
	}

	return &pb.GetFeedResponse{
		Feed: dbFeedToPb(feed),
	}, nil
}

func (s *FeedService) GetFeedByURL(ctx context.Context, req *pb.GetFeedByURLRequest) (*pb.GetFeedByURLResponse, error) {
	feed, err := s.manager.GetFeedByURL(req.Url)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "feed not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get feed: %v", err)
	}

	return &pb.GetFeedByURLResponse{
		Feed: dbFeedToPb(feed),
	}, nil
}

func (s *FeedService) ListFeeds(ctx context.Context, req *pb.ListFeedsRequest) (*pb.ListFeedsResponse, error) {
	// ListFeeds returns only visible feeds
	feeds, err := s.manager.GetVisibleFeeds()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list feeds: %v", err)
	}

	pbFeeds := make([]*pb.Feed, len(feeds))
	for i, feed := range feeds {
		pbFeeds[i] = dbFeedToPb(feed)
	}

	return &pb.ListFeedsResponse{
		Feeds: pbFeeds,
	}, nil
}

func (s *FeedService) ListAllFeeds(ctx context.Context, req *pb.ListAllFeedsRequest) (*pb.ListAllFeedsResponse, error) {
	feeds, err := s.manager.GetAllFeeds()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list feeds: %v", err)
	}

	pbFeeds := make([]*pb.Feed, len(feeds))
	for i, feed := range feeds {
		pbFeeds[i] = dbFeedToPb(feed)
	}

	return &pb.ListAllFeedsResponse{
		Feeds: pbFeeds,
	}, nil
}

func (s *FeedService) UpdateFeed(ctx context.Context, req *pb.UpdateFeedRequest) (*pb.UpdateFeedResponse, error) {
	// Low-level update - not typically used
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) UpdateFeedError(ctx context.Context, req *pb.UpdateFeedErrorRequest) (*pb.UpdateFeedErrorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) ClearFeedError(ctx context.Context, req *pb.ClearFeedErrorRequest) (*pb.ClearFeedErrorResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) DeleteFeed(ctx context.Context, req *pb.DeleteFeedRequest) (*pb.DeleteFeedResponse, error) {
	if err := s.manager.DeleteFeed(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete feed: %v", err)
	}

	return &pb.DeleteFeedResponse{}, nil
}

func (s *FeedService) HideFeed(ctx context.Context, req *pb.HideFeedRequest) (*pb.HideFeedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "use HideFeedByURL instead")
}

func (s *FeedService) ShowFeed(ctx context.Context, req *pb.ShowFeedRequest) (*pb.ShowFeedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "use ShowFeedByURL instead")
}

func (s *FeedService) HideFeedByURL(ctx context.Context, req *pb.HideFeedByURLRequest) (*pb.HideFeedByURLResponse, error) {
	if err := s.manager.HideFeedByURL(req.Url); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hide feed: %v", err)
	}

	return &pb.HideFeedByURLResponse{}, nil
}

func (s *FeedService) ShowFeedByURL(ctx context.Context, req *pb.ShowFeedByURLRequest) (*pb.ShowFeedByURLResponse, error) {
	if err := s.manager.ShowFeedByURL(req.Url); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to show feed: %v", err)
	}

	return &pb.ShowFeedByURLResponse{}, nil
}

func (s *FeedService) GetFeedStats(ctx context.Context, req *pb.GetFeedStatsRequest) (*pb.GetFeedStatsResponse, error) {
	stats, err := s.manager.GetFeedStats()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get feed stats: %v", err)
	}

	pbStats := make([]*pb.FeedStats, len(stats))
	for i, stat := range stats {
		pbStats[i] = dbFeedStatsToPb(stat)
	}

	return &pb.GetFeedStatsResponse{
		Stats: pbStats,
	}, nil
}

func (s *FeedService) SearchFeedsByTitle(ctx context.Context, req *pb.SearchFeedsByTitleRequest) (*pb.SearchFeedsByTitleResponse, error) {
	results, err := s.manager.SearchFeedsByTitle(req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search feeds: %v", err)
	}

	pbResults := make([]*pb.FeedStats, len(results))
	for i, result := range results {
		pbResults[i] = &pb.FeedStats{
			Id:          result.ID,
			Title:       result.Title,
			Url:         result.Url,
			TotalItems:  result.TotalItems,
			UnreadItems: result.UnreadItems,
		}
	}

	return &pb.SearchFeedsByTitleResponse{
		Results: pbResults,
	}, nil
}

func (s *FeedService) SearchFeedsGlobally(ctx context.Context, req *pb.SearchFeedsGloballyRequest) (*pb.SearchFeedsGloballyResponse, error) {
	results, err := s.manager.SearchFeedsGlobally(req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search feeds: %v", err)
	}

	pbResults := make([]*pb.FeedStats, len(results))
	for i, result := range results {
		pbResults[i] = &pb.FeedStats{
			Id:          result.ID,
			Title:       result.Title,
			Url:         result.Url,
			TotalItems:  result.TotalItems,
			UnreadItems: result.UnreadItems,
		}
	}

	return &pb.SearchFeedsGloballyResponse{
		Results: pbResults,
	}, nil
}

// Item operations

func (s *FeedService) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) ListItemsByFeed(ctx context.Context, req *pb.ListItemsByFeedRequest) (*pb.ListItemsByFeedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) UpsertItem(ctx context.Context, req *pb.UpsertItemRequest) (*pb.UpsertItemResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) UpsertItems(ctx context.Context, req *pb.UpsertItemsRequest) (*pb.UpsertItemsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) DeleteItemsByFeed(ctx context.Context, req *pb.DeleteItemsByFeedRequest) (*pb.DeleteItemsByFeedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) GetItemsWithReadStatus(ctx context.Context, req *pb.GetItemsWithReadStatusRequest) (*pb.GetItemsWithReadStatusResponse, error) {
	items, err := s.manager.GetItemsWithReadStatus(req.FeedId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get items: %v", err)
	}

	pbItems := make([]*pb.ItemWithReadStatus, len(items))
	for i, item := range items {
		pbItems[i] = dbItemWithReadStatusToPb(item)
	}

	return &pb.GetItemsWithReadStatusResponse{
		Items: pbItems,
	}, nil
}

func (s *FeedService) SearchItemsByTitle(ctx context.Context, req *pb.SearchItemsByTitleRequest) (*pb.SearchItemsByTitleResponse, error) {
	results, err := s.manager.SearchItemsByTitle(req.FeedId, req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search items: %v", err)
	}

	pbResults := make([]*pb.ItemWithReadStatus, len(results))
	for i, result := range results {
		pbItem := &pb.Item{
			Id:          result.ID,
			FeedId:      result.FeedID,
			Guid:        result.Guid,
			Title:       result.Title,
			Description: result.Description,
			Content:     result.Content,
			Link:        result.Link,
		}

		if result.Published.Valid {
			pbItem.Published = timestamppb.New(result.Published.Time)
		}
		if result.CreatedAt.Valid {
			pbItem.CreatedAt = timestamppb.New(result.CreatedAt.Time)
		}

		pbResults[i] = &pb.ItemWithReadStatus{
			Item: pbItem,
			Read: result.Read,
		}
	}

	return &pb.SearchItemsByTitleResponse{
		Results: pbResults,
	}, nil
}

func (s *FeedService) SearchItemsGlobally(ctx context.Context, req *pb.SearchItemsGloballyRequest) (*pb.SearchItemsGloballyResponse, error) {
	results, err := s.manager.SearchItemsGlobally(req.FeedId, req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search items: %v", err)
	}

	pbResults := make([]*pb.ItemWithReadStatus, len(results))
	for i, result := range results {
		pbItem := &pb.Item{
			Id:          result.ID,
			FeedId:      result.FeedID,
			Guid:        result.Guid,
			Title:       result.Title,
			Description: result.Description,
			Content:     result.Content,
			Link:        result.Link,
		}

		if result.Published.Valid {
			pbItem.Published = timestamppb.New(result.Published.Time)
		}
		if result.CreatedAt.Valid {
			pbItem.CreatedAt = timestamppb.New(result.CreatedAt.Time)
		}

		pbResults[i] = &pb.ItemWithReadStatus{
			Item: pbItem,
			Read: result.Read,
		}
	}

	return &pb.SearchItemsGloballyResponse{
		Results: pbResults,
	}, nil
}

// Read status operations

func (s *FeedService) MarkItemRead(ctx context.Context, req *pb.MarkItemReadRequest) (*pb.MarkItemReadResponse, error) {
	if err := s.manager.MarkItemRead(req.ItemId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark item read: %v", err)
	}

	return &pb.MarkItemReadResponse{}, nil
}

func (s *FeedService) MarkItemUnread(ctx context.Context, req *pb.MarkItemUnreadRequest) (*pb.MarkItemUnreadResponse, error) {
	if err := s.manager.MarkItemUnread(req.ItemId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark item unread: %v", err)
	}

	return &pb.MarkItemUnreadResponse{}, nil
}

func (s *FeedService) MarkAllItemsReadInFeed(ctx context.Context, req *pb.MarkAllItemsReadInFeedRequest) (*pb.MarkAllItemsReadInFeedResponse, error) {
	if err := s.manager.MarkAllItemsReadInFeed(req.FeedId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark all items read: %v", err)
	}

	return &pb.MarkAllItemsReadInFeedResponse{}, nil
}

func (s *FeedService) IsItemRead(ctx context.Context, req *pb.IsItemReadRequest) (*pb.IsItemReadResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Folder operations

func (s *FeedService) AddFeedFolder(ctx context.Context, req *pb.AddFeedFolderRequest) (*pb.AddFeedFolderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) GetFeedFolders(ctx context.Context, req *pb.GetFeedFoldersRequest) (*pb.GetFeedFoldersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) DeleteFeedFolders(ctx context.Context, req *pb.DeleteFeedFoldersRequest) (*pb.DeleteFeedFoldersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *FeedService) GetFolderStats(ctx context.Context, req *pb.GetFolderStatsRequest) (*pb.GetFolderStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Log operations

func (s *FeedService) GetLogMessages(ctx context.Context, req *pb.GetLogMessagesRequest) (*pb.GetLogMessagesResponse, error) {
	messages, err := s.manager.GetLogMessages(req.Limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get log messages: %v", err)
	}

	pbMessages := make([]*pb.LogMessage, len(messages))
	for i, msg := range messages {
		pbMessages[i] = dbLogMessageToPb(msg)
	}

	return &pb.GetLogMessagesResponse{
		Messages: pbMessages,
	}, nil
}

func (s *FeedService) GetLogMessage(ctx context.Context, req *pb.GetLogMessageRequest) (*pb.GetLogMessageResponse, error) {
	message, err := s.manager.GetLogMessage(req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "log message not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get log message: %v", err)
	}

	return &pb.GetLogMessageResponse{
		Message: dbLogMessageToPb(message),
	}, nil
}

func (s *FeedService) DeleteAllLogMessages(ctx context.Context, req *pb.DeleteAllLogMessagesRequest) (*pb.DeleteAllLogMessagesResponse, error) {
	if err := s.manager.DeleteAllLogMessages(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete log messages: %v", err)
	}

	return &pb.DeleteAllLogMessagesResponse{}, nil
}
