package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	f "github.com/kubesure/forkjoin"
)

//DispatchWorker dispatches to the configured URL
type DispatchWorker struct {
	Request               HTTPRequest
	activeDeadLineSeconds uint32
}

//Work dispatches http request and stream a response back
func (hdw *DispatchWorker) Work(ctx context.Context, x interface{}) <-chan f.Result {
	resultStream := make(chan f.Result)
	log := f.NewLogger()

	go func() {
		defer close(resultStream)

		if len(hdw.Request.Message.Method) == 0 || len(hdw.Request.Message.URL) == 0 {
			log.LogInvalidRequest(RequestID(ctx), hdw.Request.Message.ID, "Method or URL is empty")
			resultStream <- f.Result{ID: RequestID(ctx), X: makeErrorResponse(hdw.Request, http.StatusBadRequest),
				Err: &f.FJError{Code: f.RequestError, Message: "Method or URL is empty"}}
			return
		}

		var validMethod bool = false
		var methods = []METHOD{GET, POST, PUT, PATCH}

		for _, method := range methods {
			if hdw.Request.Message.Method == method {
				validMethod = true
				break
			}
		}

		if !validMethod {
			log.LogInvalidRequest(RequestID(ctx), hdw.Request.Message.ID, fmt.Sprintf("Method %v is invalid", hdw.Request.Message.Method))
			resultStream <- f.Result{ID: RequestID(ctx), X: makeErrorResponse(hdw.Request, http.StatusBadRequest),
				Err: &f.FJError{Code: f.RequestError, Message: fmt.Sprintf("Method %v is invalid", hdw.Request.Message.Method)}}
			return
		}
		httpDispatch(ctx, hdw.Request, resultStream)
	}()
	return resultStream
}

func httpDispatch(ctx context.Context, reqMsg HTTPRequest, resultStream chan<- f.Result) {
	responseStream := make(chan f.Result)
	log := f.NewLogger()

	go func() {
		defer close(responseStream)
		req, _ := http.NewRequestWithContext(
			ctx, string(reqMsg.Message.Method),
			reqMsg.Message.URL,
			strings.NewReader(reqMsg.Message.Payload))

		for k, v := range reqMsg.Message.Headers {
			req.Header.Add(k, v)
		}

		var client *http.Client

		// Append the certificates from the CA
		client = newFunction(reqMsg, client, req, log)

		res, err := client.Do(req)

		if ctx.Err() != nil && res == nil {
			log.LogAbortedRequest(RequestID(ctx), reqMsg.Message.ID, "Aborted")
			responseStream <- f.Result{ID: RequestID(ctx), X: makeErrorResponse(reqMsg, http.StatusRequestTimeout),
				Err: &f.FJError{Code: f.RequestAborted, Message: fmt.Sprintf("Request aborted took longer than expected %v", err)}}
		} else if err != nil {
			log.LogRequestDispatchError(RequestID(ctx), reqMsg.Message.ID, err.Error())
			responseStream <- f.Result{ID: RequestID(ctx), X: makeErrorResponse(reqMsg, http.StatusBadGateway),
				Err: &f.FJError{Code: f.ConnectionError, Message: fmt.Sprintf("Error in dispatching request: %v", err)}}
		} else {
			bb, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.LogResponseError(RequestID(ctx), reqMsg.Message.ID, err.Error())
				responseStream <- f.Result{ID: RequestID(ctx), X: makeResponse(reqMsg, res, nil),
					Err: &f.FJError{Code: f.ResponseError, Message: fmt.Sprintf("Error reading http response: %v", err)}}
			}

			hr := makeResponse(reqMsg, res, bb)

			for k, values := range res.Header {
				for _, value := range values {
					hr.Message.Add(k, value)
				}
			}
			defer res.Body.Close()
			responseStream <- f.Result{ID: RequestID(ctx), X: hr}
		}
	}()

	// TODO: make this blocking
	for {
		select {
		case r := <-responseStream:
			resultStream <- r
			return
		default:
		}
	}
}

func newFunction(reqMsg HTTPRequest, client *http.Client, req *http.Request, log *f.StandardLogger) *http.Client {
	if reqMsg.Message.Authentication == NONE {
		client = &http.Client{}
	} else if reqMsg.Message.Authentication == BASIC {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(reqMsg.Message.BasicAtuhCredentials.ServerCertificate))
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			},
		}
		req.SetBasicAuth(reqMsg.Message.BasicAtuhCredentials.UserName,
			reqMsg.Message.BasicAtuhCredentials.Password)
	} else if reqMsg.Message.Authentication == MUTUAL {
		certificate, err := tls.LoadX509KeyPair("client_CRT", "client_KEY")
		if err != nil {
			log.Fatalf("could not load client key pair: %s", err)
		}

		caCertPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile("CA_CRT")
		if err != nil {
			log.Fatalf("could not read ca certificate: %s", err)
		}

		if ok := caCertPool.AppendCertsFromPEM(ca); !ok {
			log.Fatalf("failed to append ca certs")
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      caCertPool,
					Certificates: []tls.Certificate{certificate},
				},
			},
		}
	}
	return client
}

func makeResponse(reqMsg HTTPRequest, res *http.Response, bb []byte) HTTPResponse {
	hr := HTTPResponse{
		Message: HTTPMessage{
			ID:             reqMsg.Message.ID,
			StatusCode:     res.StatusCode,
			Method:         reqMsg.Message.Method,
			URL:            reqMsg.Message.URL,
			Payload:        string(bb),
			ActiveDeadLine: reqMsg.Message.ActiveDeadLine,
		},
	}
	return hr
}

func makeErrorResponse(reqMsg HTTPRequest, resStatus int) HTTPResponse {
	hr := HTTPResponse{
		Message: HTTPMessage{
			ID:             reqMsg.Message.ID,
			StatusCode:     resStatus,
			Method:         reqMsg.Message.Method,
			URL:            reqMsg.Message.URL,
			Payload:        fmt.Sprintf(""),
			ActiveDeadLine: reqMsg.Message.ActiveDeadLine,
		},
	}
	return hr
}

func (hdw *DispatchWorker) ActiveDeadLineSeconds() uint32 {
	return hdw.activeDeadLineSeconds
}
