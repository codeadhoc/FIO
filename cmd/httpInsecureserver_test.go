package main

import (
	"context"
	"io"
	"log"
	"os"
	"testing"

	h "github.com/kubesure/forkjoin/http"
	"google.golang.org/grpc"
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

func TestHTTPFIO(t *testing.T) {
	conn := makeGrpcConn()
	defer conn.Close()
	c := h.NewHTTPForkJoinServiceClient(conn)

	req := h.Request{Id: "BIN1", Messages: makeValidRequests()}
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
		res = append(res, response)
	}
	if len(res) != 2 {
		t.Error("there should be a 2 responses")
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

func TestSlowHTTPResponse(t *testing.T) {
	conn := makeGrpcConn()
	defer conn.Close()
	c := h.NewHTTPForkJoinServiceClient(conn)

	req := h.Request{Messages: makeSlowResponseRequests()}
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

	if len(res) != 1 {
		t.Error("there should be a 1 errors")
	}

	log.Printf("code %v", res[0].Errors[0].Code)

	if res[0].Errors[0].Code != h.ErrorCode_ConnectionError {
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

func makeGrpcConn() *grpc.ClientConn {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}

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

func makeSlowResponseRequests() []*h.Message {
	delayedMesg := &h.Message{Method: h.Message_POST, URL: "http://localhost:8000/healthz"}
	msgs := []*h.Message{}
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
