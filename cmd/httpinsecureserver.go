package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	h "github.com/kubesure/forkjoin/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime)
}

const (
	port = ":50051"
)

func main() {
	log.Println("starting http fork join service...")
	ctx := context.Background()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	server := &h.DispatchServer{}
	h.RegisterHTTPForkJoinServiceServer(s, server)
	reflection.Register(s)

	h := health.NewServer()
	h.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, h)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Println("shutting down fork join service...")
			s.GracefulStop()
			<-ctx.Done()
		}
	}()
	s.Serve(lis)
}
