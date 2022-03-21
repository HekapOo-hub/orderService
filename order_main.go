package main

import (
	"context"
	"github.com/HekapOo-hub/orderService/internal/handler"
	"github.com/HekapOo-hub/orderService/internal/proto/orderpb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

func main() {
	orderHandler, err := handler.NewOrderHandler(context.Background())
	if err != nil {
		log.Warnf("%v", err)
		return
	}
	lis, err := net.Listen("tcp", handler.OrderPort)
	if err != nil {
		log.Warnf("error %v", err)
		return
	}
	server := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(server, orderHandler)
	if err := server.Serve(lis); err != nil {
		log.Warnf("%v", err)
	}
}
