package api

import (
	"context"
	"grpc/gateway/pb"
	"math/rand/v2"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedUserServiceServer
}

func (s *Server) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.UserResponse, error) {

	succesRand := rand.IntN(2)
	if succesRand == 1 {
		return &pb.UserResponse{
			User: &pb.User{
				Id:    req.Id,
				Name:  "Rakhasa",
				Email: "rakhasa@mail.com",
			},
		}, nil
	}

	return nil, status.Error(codes.Code(http.StatusNotFound), "user not found")
}

func (s *Server) CreateUser(
	ctx context.Context,
	req *pb.CreateUserRequest,
) (*pb.UserResponse, error) {

	return &pb.UserResponse{
		User: &pb.User{
			Id:    1,
			Name:  req.Name,
			Email: req.Email,
		},
	}, nil
}

func (s *Server) DeleteUser(
	ctx context.Context,
	req *pb.DeleteUserRequest,
) (*pb.DeleteUserResponse, error) {

	return &pb.DeleteUserResponse{
		Message: "deleted",
	}, nil
}
