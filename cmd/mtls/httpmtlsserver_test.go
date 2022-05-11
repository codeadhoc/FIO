package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	h "github.com/kubesure/forkjoin/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime)
}

const (
	address = "localhost:50051"
)

//run ../test/test_server.go
//run httpfjserver.go
//TODO: Create mock GRPC tests

func TestHTTPForkJoin(t *testing.T) {
	conn := makeGrpcConn()
	defer conn.Close()
	c := h.NewHTTPForkJoinServiceClient(conn)

	req := h.Request{Messages: makeValidRequests()}
	stream, err := c.FanoutFanin(context.Background(), &req)
	if err != nil {
		t.Errorf("GRPC error call should have not failed with %v", err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("%v.FanoutFanin = _, %v", c, err)
		}

		if response.Message.StatusCode != 200 {
			t.Errorf("error code is not 200 but %v", response.Message.StatusCode)
		}

		if response.Message.URL != "https://httpbin.org/anything" {
			t.Errorf("URL not found in response")
		}
		if response.Message.Method == h.Message_NIL {
			t.Errorf("response method is empty")
		}
		if len(response.Message.Payload) == 0 {
			t.Errorf("response payload is empty")
		}
		if len(response.Message.Headers) == 0 {
			t.Errorf("response headers is empty")
		}
	}
}

func TestInvalidHTTPURLForkJoin(t *testing.T) {
	conn := makeGrpcConn()
	defer conn.Close()
	c := h.NewHTTPForkJoinServiceClient(conn)

	req := h.Request{Messages: makeInValidURLRequests()}
	stream, err := c.FanoutFanin(context.Background(), &req)
	if err != nil {
		t.Errorf("GRPC error call should have not failed with %v", err)
	}

	res := []*h.Response{}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("%v.FanoutFanin = _, %v", c, err)
		}
		res = append(res, response)
	}

	if len(res) != 2 {
		t.Error("there should be a 2 errors")
	}

	if res[0].Errors[0].Code != h.ErrorCode_ConnectionError {
		t.Errorf("there should be error code: %v", h.ErrorCode_ConnectionError)
	}

	if res[1].Errors[0].Code != h.ErrorCode_ConcurrencyContextError {
		t.Errorf("there should be error code: %v", h.ErrorCode_ConnectionError)
	}
}

func TestInvalidRequestsCfgHTTPForkJoin(t *testing.T) {
	conn := makeGrpcConn()
	defer conn.Close()
	c := h.NewHTTPForkJoinServiceClient(conn)

	req := h.Request{Messages: makeInValidRequestsCfg()}
	stream, err := c.FanoutFanin(context.Background(), &req)
	if err != nil {
		t.Errorf("GRPC error call should have not failed with %v", err)
	}

	res := []*h.Response{}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("%v.FanoutFanin = _, %v", c, err)
		}
		res = append(res, response)
	}

	if len(res) == 0 {
		t.Error("there should be a 2 errors")
	}

	if res[0].Errors[0].Code != h.ErrorCode_RequestError {
		t.Errorf("there should be error code: %v", h.ErrorCode_RequestError)
	}

	if res[1].Errors[0].Code != h.ErrorCode_ConcurrencyContextError {
		t.Errorf("there should be error code: %v", h.ErrorCode_ConnectionError)
	}

}

// TODO: FIX certs passed
func makeGrpcConn() *grpc.ClientConn {
	certificate, err := tls.LoadX509KeyPair(os.Getenv("SERVER_CRT"), os.Getenv("SERVER_KEY"))
	if err != nil {
		log.Fatalf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(os.Getenv("CA_CRT"))
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:   os.Getenv("SERVER_HOSTNAME"),
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
	}

	conn, err := grpc.Dial(address, opts...)
	return conn
}

/*func makeGrpcConn() *grpc.ClientConn {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}*/

func makeValidRequests() []*h.Message {

	headers := make(map[string]string)
	headers["header1"] = "value1"

	m1 := &h.Message{Method: h.Message_GET, URL: "https://httpbin.org/anything"}
	m1.Headers = headers
	m1.Payload = "{body:'body'}"

	m2 := &h.Message{Method: h.Message_POST, URL: "https://httpbin.org/anything"}
	m2.Headers = headers
	m2.Payload = "{body:'body'}"

	msgs := []*h.Message{}
	msgs = append(msgs, m1)
	msgs = append(msgs, m2)
	return msgs
}

func makeInValidURLRequests() []*h.Message {
	invalidURLMsg := &h.Message{Method: h.Message_GET, URL: "https://unknown/anything"}
	delayedMesg := &h.Message{Method: h.Message_POST, URL: "http://localhost:8000/healthz"}
	msgs := []*h.Message{}
	msgs = append(msgs, invalidURLMsg)
	msgs = append(msgs, delayedMesg)
	return msgs
}

func makeInValidRequestsCfg() []*h.Message {
	invalidURLMsg := &h.Message{Method: h.Message_GET}
	delayedMesg := &h.Message{Method: h.Message_GET, URL: "http://localhost:8000/healthz"}
	msgs := []*h.Message{}
	msgs = append(msgs, invalidURLMsg)
	msgs = append(msgs, delayedMesg)
	return msgs
}
