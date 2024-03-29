package api

import "github.com/jcdotter/go/data"

// ----------------------------------------------------------------------------
// API TYPES

// Protocol is the protocol used to
// communicate with the API
type Protocol byte

const (
	REST Protocol = iota
	GRPC
	WEBSOCKET
	SOAP
)

// ----------------------------------------------------------------------------
// METHOD TYPES

// MethodType is the HTTP method
type MethodType byte

const (
	GET MethodType = iota
	POST
	PUT
	DELETE
	PATCH
)

var (
	methodIndex  = []int{0, 4, 9, 13, 20}
	methodString = "GET POST PUT DELETE PATCH "
	methodType   = map[string]MethodType{
		"GET":    GET,
		"POST":   POST,
		"PUT":    PUT,
		"DELETE": DELETE,
		"PATCH":  PATCH,
	}
)

func Methodtype(s string) MethodType {
	return methodType[s]
}

func (t MethodType) String() string {
	return methodString[methodIndex[t]:methodIndex[t+1]]
}

// ----------------------------------------------------------------------------
// CONTENT TYPES

// ContentType is the type of the
// content in the request or response
type ContentType byte

const (
	JSON ContentType = iota
	XML
	FORM
	TEXT
	CSV
	HTML
	XLS
	DOC
)

var (
	contentString = map[ContentType]string{
		JSON: "application/json",
		XML:  "application/xml",
		FORM: "application/x-www-form-urlencoded",
		TEXT: "text/plain",
		CSV:  "text/csv",
		HTML: "text/html",
		XLS:  "application/vnd.ms-excel",
		DOC:  "application/msword",
	}
	contentType = map[string]ContentType{
		"application/json":                  JSON,
		"application/xml":                   XML,
		"application/x-www-form-urlencoded": FORM,
		"text/plain":                        TEXT,
		"text/csv":                          CSV,
		"text/html":                         HTML,
		"application/vnd.ms-excel":          XLS,
		"application/msword":                DOC,
	}
)

func Contenttype(s string) ContentType {
	return contentType[s]
}

func (t ContentType) String() string {
	return contentString[t]
}

// ----------------------------------------------------------------------------
// API
// api.endpoint.method.request

type Api struct {
	Protocol  Protocol
	Auth      *Api
	Url       string
	endpoints endpoints
}

type endpoints struct{ *data.Data }

func Endpoints(e ...*Endpoint) endpoints {
	return endpoints{data.Of(e...)}
}

func (e *endpoints) Index(i int) *Endpoint {
	return e.Data.Index(i).(*Endpoint)
}

func (e *endpoints) Get(key string) *Endpoint {
	return e.Data.Get(key).(*Endpoint)
}

func (e *endpoints) Add(value *Endpoint) *endpoints {
	e.Data.Add(value)
	return e
}

type Endpoint struct {
	Uri     string
	Content ContentType
	Allow   ContentType
	Methods Methods
}

func (e *Endpoint) Key() string {
	return e.Uri
}

type Methods struct{ *data.Data }

func (m *Methods) Index(i int) *Method {
	return m.Data.Index(i).(*Method)
}

func (m *Methods) Get(key string) *Method {
	return m.Data.Get(key).(*Method)
}

func (m *Methods) Add(value *Method) *Methods {
	m.Data.Add(value)
	return m
}

type Method struct {
	Type     MethodType
	Request  *Request
	Response *Response
}

func (m *Method) Key() string {
	return m.Type.String()
}

type Request struct {
	Path   *data.Data
	Params *data.Data
	Header *data.Data
	Body   *data.Data
	// add webhooks
}

type Response struct {
	Header *data.Data
	Body   *data.Data
}

type param struct {
	key *string
	typ *string
}

func Param(k, t string) *param {
	return &param{
		key: &k,
		typ: &t,
	}
}

func (p *param) Key() string {
	return *p.key
}
