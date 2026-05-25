package server

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/middelmatigheid/top-searches/internal/service"
	proto "github.com/middelmatigheid/top-searches/proto"
)

// Searches

type SearchServer struct {
	proto.UnimplementedSearchesServer
	service *service.Service
}

func NewSearchServer(service *service.Service) *SearchServer {
	return &SearchServer{service: service}
}

func (s *SearchServer) GetTopN(ctx context.Context, req *proto.GetTopNRequest) (*proto.GetTopNResponse, error) {
	if req.N <= 0 {
		return nil, status.Error(codes.InvalidArgument, "N must be positive")
	}

	top, err := s.service.GetTopN(req.N)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoTop := make([]*proto.TopSearch, len(top))
	for i, item := range top {
		protoTop[i] = &proto.TopSearch{
			Search:       item.Search,
			Count:        item.Count,
			LastSearched: item.LastSearched.Format(time.RFC3339),
		}
	}

	return &proto.GetTopNResponse{TopSearches: protoTop}, nil
}

// Stoplist

type StoplistServer struct {
	proto.UnimplementedStoplistServer
	service *service.Service
}

func NewStoplistServer(service *service.Service) *StoplistServer {
	return &StoplistServer{service: service}
}

func (s *StoplistServer) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
	if req.Word == "" {
		return nil, status.Error(codes.InvalidArgument, "word is required")
	}
	s.service.AddStoplistWord(req.Word)
	return &proto.AddResponse{Success: true, Message: "Word added"}, nil
}

func (s *StoplistServer) Remove(ctx context.Context, req *proto.RemoveRequest) (*proto.RemoveResponse, error) {
	if req.Word == "" {
		return nil, status.Error(codes.InvalidArgument, "word is required")
	}
	s.service.RemoveStoplistWord(req.Word)
	return &proto.RemoveResponse{Success: true, Message: "Word removed"}, nil
}

func (s *StoplistServer) GetStoplist(ctx context.Context, req *proto.GetStoplistRequest) (*proto.GetStoplistResponse, error) {
	words := s.service.GetStoplist()
	return &proto.GetStoplistResponse{Words: words}, nil
}

func (s *StoplistServer) Contains(ctx context.Context, req *proto.ContainsRequest) (*proto.ContainsResponse, error) {
	contains := s.service.StoplistContains(req.Word)
	return &proto.ContainsResponse{Contains: contains}, nil
}
