package api

import (
	"github.com/jcdotter/go/data"
)

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
// DATA TYPES

// DataType is the type of a data element
// in the request or response
type DataType byte

const (
	BOOL DataType = iota
	INT
	FLOAT
	STRING
)

type Object map[string]any
type List []any

// ----------------------------------------------------------------------------
// API
// api.endpoint.method.request

type Api struct {
	Protocol  Protocol
	Auth      *Api
	Url       string
	Resources *data.Data
	Params    *data.Data
	Header    *data.Data
	Data      *data.Data
}

func (a *Api) Resource(key string) *Resource {
	var el any
	if el := a.Resources.Get(key); el == nil {
		return nil
	}
	r := el.(*Resource)
	return &Resource{
		Uri:       key,
		Methods:   r.Methods,
		Resources: r.Resources,
	}
}

type Resource struct {
	Uri       string
	Methods   *data.Data
	Resources *data.Data
	Params    *data.Data
	Header    *data.Data
	Data      *data.Data
}

func (r *Resource) Key() string {
	return r.Uri
}

func (r *Resource) Resource(id, name string) *Resource {
	s := r.Resources.Get(name).(*Resource)
	return &Resource{
		Uri:       id + "/" + name,
		Methods:   s.Methods,
		Resources: s.Resources,
	}
}

func (r *Resource) Method(key string) *Method {
	return r.Methods.Get(key).(*Method)
}

func (r *Resource) Get()    {}
func (r *Resource) Post()   {}
func (r *Resource) Put()    {}
func (r *Resource) Delete() {}

type Method struct {
	Name     string
	Request  *Request
	Response *Response
}

func (m *Method) Key() string {
	return m.Name
}

type Request struct {
	Path   *data.Data
	Params *data.Data
	Header *data.Data
	Body   any
	// add webhooks
}

type Response struct {
	Header *data.Data
	Body   any
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
