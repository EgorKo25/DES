package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	pb "github.com/EgorKo25/DES/internal/server/extension-service-gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ch chan chan []byte

type ExtServer struct {
	pb.UnimplementedUserExtensionServiceServer
	chanLength    int
	chanMaxLength int
}

func NewExtServer(channel chan chan []byte) *ExtServer {
	ch = channel
	return &ExtServer{chanMaxLength: cap(ch)}
}

func (es *ExtServer) GetUserExtension(ctx context.Context, in *pb.GetRequest) (out *pb.GetResponse, err error) {
	resultChan := make(chan []byte)

	data, err := json.Marshal(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "BAD REQUEST")
	}

	ch <- resultChan
	resultChan <- data

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, strings.ToUpper(ctx.Err().Error()))

	case ext := <-resultChan:
		if err = json.Unmarshal(ext, &out); err != nil {
			return nil, status.Error(codes.Internal, "INTERNAL SERVER ERROR")
		}
		return out, status.Error(codes.OK, "OK")
	}
}

func (es *ExtServer) LogUnaryRpcInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

	before := time.Now()

	m, err := handler(ctx, req)
	_ = time.Since(before)

	return m, err

}

func (es *ExtServer) StartServer(port string) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", port)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to listen: %v", err))
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(es.LogUnaryRpcInterceptor),
	)
	pb.RegisterUserExtensionServiceServer(s, &ExtServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return s, nil
}
