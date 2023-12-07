package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"

	pb "github.com/EgorKo25/DES/internal/server/extension-service-gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ch chan chan []byte

type ExtServer struct {
	pb.UnimplementedUserExtensionServiceServer

	logger     *zap.Logger
	grpcLogger *zap.Logger
}

func NewExtServer(channel chan chan []byte, logger, grpcLogger *zap.Logger) *ExtServer {
	ch = channel
	return &ExtServer{
		logger:     logger,
		grpcLogger: grpcLogger,
	}
}

func (es *ExtServer) GetUserExtension(ctx context.Context, in *pb.GetRequest) (out *pb.GetResponse, err error) {
	resultChan := make(chan []byte)

	data, err := json.Marshal(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "BAD REQUEST")
	}

	select {
	case ch <- resultChan:
		resultChan <- data
	default:
		log.Println(len(ch), cap(ch))
		return nil, status.Error(codes.ResourceExhausted, "TOO MANY REQUEST")
	}
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

func (es *ExtServer) LogUnaryRPCInterceptor(ctx context.Context, req interface{},
	_ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	timeStart := time.Now()

	es.grpcLogger.Info(
		"grpc connection start",
		zap.String("grpc request", req.(*pb.GetRequest).String()),
	)

	m, err := handler(ctx, req)

	isRespNil := resp != nil

	es.grpcLogger.Info(
		"grpc connection done",
		zap.Bool("is grpc response", isRespNil),
		zap.NamedError("response error", err),
		zap.Duration("duration", time.Since(timeStart)),
	)
	return m, err
}

func (es *ExtServer) StartServer(port string) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(es.LogUnaryRPCInterceptor),
	)
	pb.RegisterUserExtensionServiceServer(s, &ExtServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			es.logger.Fatal("failed to serve",
				zap.String("error", err.Error()),
			)
		}
	}()

	return s, nil
}
