package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"

	h "github.com/kubesure/forkjoin/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var crtFile = os.Getenv("SERVER_CRT")
var keyFile = os.Getenv("SERVER_KEY")

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

	certificate, err := tls.LoadX509KeyPair(os.Getenv("SERVER_CRT"), os.Getenv("SERVER_KEY"))
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(os.Getenv("CA_CRT"))
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append client certs")
	}

	opts := []grpc.ServerOption{
		grpc.Creds(
			credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{certificate},
				ClientCAs:    certPool,
			},
			)),
	}

	s := grpc.NewServer(opts...)
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
