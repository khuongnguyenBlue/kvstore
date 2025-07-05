package server

import (
	"context"
	"kvstore/internal/storage"
	pb "kvstore/pkg/pb/api/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedKVStoreServer
	storage storage.Storage
}

func New(storage storage.Storage) *Server {
	return &Server{storage: storage}
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "key cannot be empty")
	}

	val, found := s.storage.Get(req.GetKey())
	return &pb.GetResponse{
		Value: val,
		Found: found,
	}, nil
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if req.GetKey() == "" {
		return &pb.SetResponse{Success: false}, status.Error(codes.InvalidArgument, "key cannot be empty")
	}

	err := s.storage.Set(req.GetKey(), req.GetValue())
	if err != nil {
		return &pb.SetResponse{Success: false},
			status.Errorf(codes.Internal, "failed to set key: %v", err)
	}

	return &pb.SetResponse{
		Success: true,
	}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
    if req.GetKey() == "" {
        return nil, status.Error(codes.InvalidArgument, "key cannot be empty")
    }
    
    existed, err := s.storage.Delete(req.GetKey())
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to delete key: %v", err)
    }
    
    return &pb.DeleteResponse{
        Success: true,
        Existed: existed,
    }, nil
}

func (s *Server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
    limit := int(req.GetLimit())
    if limit < 0 {
        return nil, status.Error(codes.InvalidArgument, "limit cannot be negative")
    }
    
    data, err := s.storage.List(limit)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to list keys: %v", err)
    }
    
    var pairs []*pb.KeyValuePair
    for k, v := range data {
        pairs = append(pairs, &pb.KeyValuePair{
            Key:   k,
            Value: v,
        })
    }
    
    return &pb.ListResponse{Pairs: pairs}, nil
}
