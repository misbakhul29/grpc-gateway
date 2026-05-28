package main

import (
	"grpc/gateway/internal/config"
	"grpc/gateway/internal/logger"
	"grpc/gateway/internal/services/api"
	"grpc/gateway/pb"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":"+config.EnvCfg.UserPort)
	if err != nil {
		logger.Log.Fatal("user-service", err.Error())
	}

	grpcServer := grpc.NewServer()

	pb.RegisterUserServiceServer(
		grpcServer,
		&api.Server{},
	)

	logger.Log.Info("user-service", "user-service running :"+config.EnvCfg.UserPort)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("user-service", err.Error())
	}
}
