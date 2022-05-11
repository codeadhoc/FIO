package forkjoin

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger() *StandardLogger {
	baseLogger := logrus.New()
	baseLogger.SetOutput(os.Stdout)
	baseLogger.SetLevel(logrus.InfoLevel)
	var sl = &StandardLogger{baseLogger}
	sl.Formatter = &logrus.JSONFormatter{}
	return sl
}

var (
	internalError        = LogEvent{InternalError, "Request Id: %s Message Id: %s Internal Error: %v"}
	invalidRequest       = LogEvent{RequestError, "Request Id: %s Message Id: Invalid Request: %v"}
	requestDispatchError = LogEvent{RequestError, "Request Id: %s Message Id: %s Request Dispatch Error: %v"}
	abortRequest         = LogEvent{RequestAborted, "Request Id: %s Message Id: %s Abort Request: %v"}
	responseError        = LogEvent{ResponseError, "Request Id: %s Message Id: %s Response error: %v"}
	connectionError      = LogEvent{ConnectionError, "Request Id: %s Message Id: %s Connection error: %v"}
	infoRequest          = LogEvent{RequestInfo, "Request Id: %s Message Id: %s: %s"}
	info                 = LogEvent{Info, "Request Id: %s %s "}
)

func (l *StandardLogger) LogInvalidRequest(requestID, messageID, message string) {
	l.Errorf(invalidRequest.message, requestID, messageID, message)
}

func (l *StandardLogger) LogResponseError(requestID, messageID, message string) {
	l.Errorf(responseError.message, requestID, messageID, message)
}

func (l *StandardLogger) LogRequestDispatchError(requestID, messageID, message string) {
	l.Errorf(requestDispatchError.message, requestID, messageID, message)
}

func (l *StandardLogger) LogAbortedRequest(requestID, messageID, message string) {
	l.Errorf(abortRequest.message, requestID, messageID, message)
}

//TODO: have one info or request and one generic
func (l *StandardLogger) LogRequestInfo(requestID, message string) {
	l.Infof(infoRequest.message, requestID, message)
}

func (l *StandardLogger) LogInfo(requestID, message string) {
	l.Infof(info.message, requestID, message)
}
