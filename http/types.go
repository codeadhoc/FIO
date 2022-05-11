package http

//METHOD http methods supported by http dispatcher
type METHOD string

//METHOD http methods supported by http dispatcher
const (
	GET   METHOD = "GET"
	POST  METHOD = "POST"
	PUT   METHOD = "PUT"
	PATCH METHOD = "PATCH"
)

// TODO: Is this type required?
type AUTHENTICATION string

const (
	NONE   AUTHENTICATION = "NONE"
	BASIC  AUTHENTICATION = "BASIC"
	MUTUAL AUTHENTICATION = "MUTUAL"
)

type BasicAuthCredentials struct {
	UserName          string
	Password          string
	ServerCertificate string
}

type MutualAuthCredentials struct {
	ClientCertificate string
	ClientKey         string
	CACertificate     string
}

//HTTPRequest URL and method to be dispatched too
type HTTPRequest struct {
	ID      string
	Message HTTPMessage
}

//HTTPResponse URL and method to be dispatched too
type HTTPResponse struct {
	Message HTTPMessage
}

//HTTPMessage URL and method to be dispatched too
type HTTPMessage struct {
	ID                    string
	URL                   string
	Method                METHOD
	Payload               string
	Headers               map[string]string
	ActiveDeadLine        uint32
	StatusCode            int
	Authentication        AUTHENTICATION
	BasicAtuhCredentials  BasicAuthCredentials
	MutualAuthCredentials MutualAuthCredentials
}

//Add adds headers to messsage
func (hm *HTTPMessage) Add(key, value string) {
	if hm.Headers == nil {
		hm.Headers = make(map[string]string)
	}
	hm.Headers[key] = value
}
