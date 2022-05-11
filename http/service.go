package http

import (
	"context"
	"fmt"

	fj "github.com/kubesure/forkjoin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//Request ID key passsed as data in context
type ctxKey int

const (
	CtxRequestID ctxKey = iota
)

//DispatchServer implements GRPC server interface FannoutFannin
type DispatchServer struct {
	UnimplementedHTTPForkJoinServiceServer
}

// TODO: Validate Input
//FanoutFanin Fans out each http message to http dispatch works using the fork join interface
func (s *DispatchServer) FanoutFanin(request *Request, stream HTTPForkJoinService_FanoutFaninServer) error {
	ctx := context.WithValue(context.Background(), CtxRequestID, request.Id)
	log := fj.NewLogger()
	mtplx := fj.NewMultiplexer()
	for i, m := range request.Messages {
		method := Message_Method_name[int32(m.Method)]
		auth := Message_Authentication_name[int32(m.Authentication)]

		// TODO: use makemessage func
		msg := HTTPMessage{
			Method:         METHOD(method),
			URL:            m.URL,
			Payload:        m.Payload,
			Headers:        m.Headers,
			ActiveDeadLine: m.ActiveDeadLineSeconds,
			ID:             fmt.Sprint(i + 1),
			Authentication: AUTHENTICATION(auth),
		}

		if auth == string(BASIC) {
			msg.BasicAtuhCredentials = BasicAuthCredentials{UserName: m.BasicAuthcredentials.UserName,
				Password:          m.BasicAuthcredentials.Password,
				ServerCertificate: m.BasicAuthcredentials.ServerCertificate}
		} else if auth == string(MUTUAL) {
			msg.MutualAuthCredentials = MutualAuthCredentials{
				ClientCertificate: m.MutualAuthCredentials.ClientCertificate,
				ClientKey:         m.MutualAuthCredentials.ClientKey,
				CACertificate:     m.MutualAuthCredentials.CACertificate,
			}
		}

		reqMsg := HTTPRequest{Message: msg}
		mtplx.AddWorker(&DispatchWorker{Request: reqMsg, activeDeadLineSeconds: msg.ActiveDeadLine})
	}

	resultStream := mtplx.Multiplex(ctx, nil)
	log.LogInfo(RequestID(ctx), "Forked")
	for result := range resultStream {
		response, _ := result.X.(HTTPResponse)
		if result.Err != nil {
			// TODO: should it return or process next message
			err := stream.Send(makeGrpcErrRes(result.ID, *result.Err, response))
			if err != nil {
				log.LogResponseError(result.ID, response.Message.ID, fmt.Sprintf("Error while writing to stream: %v", err.Error()))
				return status.Errorf(codes.Internal, fmt.Sprintf("Error while request id: %s message id: %s writing to stream", result.ID, response.Message.ID), err)
			}
		} else {
			m := makeErrResMsg(response)
			r := Response{Id: result.ID, Message: m}
			err := stream.Send(&r)
			if err != nil {
				log.LogResponseError(result.ID, response.Message.ID, fmt.Sprintf("Error while writing to stream: %v", err.Error()))
				return status.Errorf(codes.Internal, fmt.Sprintf("Error while request id: %s message id: %s writing to stream", result.ID, response.Message.ID), err)
			}
		}
	}
	log.LogInfo(RequestID(ctx), "Joined")
	return nil
}

func makeErrResMsg(response HTTPResponse) *Message {
	method := Message_Method_value[string(response.Message.Method)]
	m := &Message{
		URL:        response.Message.URL,
		Method:     Message_Method(method),
		Headers:    response.Message.Headers,
		Payload:    response.Message.Payload,
		StatusCode: uint32(response.Message.StatusCode),
		Id:         response.Message.ID,
	}
	return m
}

//func makeGrpcErrRes(code fj.EventCode, msg string) *HTTPResponse {
func makeGrpcErrRes(reqID string, err fj.FJError, response HTTPResponse) *Response {
	fjerrors := []*Error{}
	fjerr := Error{Code: ErrorCode(err.Code), Message: err.Message}
	fjerrors = append(fjerrors, &fjerr)
	m := makeErrResMsg(response)
	r := Response{Id: reqID, Message: m, Errors: fjerrors}
	return &r
}

func RequestID(ctx context.Context) string {
	return ctx.Value(CtxRequestID).(string)
}
