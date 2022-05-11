### GRPC server side result stream

GRPC server side streaming interface fans out multiple http requests to forkjoin base library using HTTP DispatchWorker as worker. DispatchWorker returns response or error as result to Result chan. GRPC server streams response to GRPC client as concurrent processing is completed for each HTTP URL requested. 

set ENV var TLS and MLTS server

1. SERVER_CRT path to server certificate. [httpMTLSServer.go](./cmd/mtls/httpMTLSServer.go) [httpTLSServer.go](./cmd/tls/httpTLSServer.go)
2. SERVER_KEY path to server key. [httpMTLSServer.go](./cmd/mtls/httpMTLSServer.go) [httpTLSServer.go](./cmd/tls/httpTLSServer.go) 
4. CA_CRT path ca certificate. required by MTLS connection refer [httpMTLSServer.go](./cmd/mtls/httpMTLSServer.go)
5. SERVER_HOSTNAME DNS name specified in certificate file. Required by clients. refer [certs.md](./certs/certs.md)

### test case insecure server 

```
func TestHTTPForkJoin(t *testing.T) {
	conn := makeGrpcConn()
	defer conn.Close()
	c := h.NewHTTPForkJoinServiceClient(conn)

	req := h.HTTPRequest{Messages: makeValidRequests()}
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
```