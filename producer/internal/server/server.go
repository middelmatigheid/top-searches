package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/middelmatigheid/top-searches/producer/internal/models"
	"github.com/middelmatigheid/top-searches/producer/internal/service"
	proto "github.com/middelmatigheid/top-searches/producer/proto"
)

type Service interface {
	SendSearch(request models.Search) error
}

type SearchServer struct {
	proto.UnimplementedProducerServer
	service Service
}

func NewSearchServer(service *service.Service) *SearchServer {
	return &SearchServer{service: service}
}

func (s *SearchServer) SendSearch(ctx context.Context, req *proto.SendSearchRequest) (*proto.SendSearchResponse, error) {
	search := models.Search{
		Search: req.Search,
		User:   req.User,
	}
	err := s.service.SendSearch(search)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.SendSearchResponse{Success: true, Message: "Search sent successfully"}, nil
}
