package api

import (
	"strconv"

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
	NONE DataType = iota
	BOOL
	INT
	FLOAT
	STRING
	LIST
	OBJECT
)

var (
	dataTypeString = "noneboolintfloatstringlistobject"
	dataTypeIndex  = []int{0, 4, 8, 11, 16, 22, 26, 32}
	dataType       = map[string]DataType{
		"none":   NONE,
		"bool":   BOOL,
		"int":    INT,
		"float":  FLOAT,
		"string": STRING,
		"list":   LIST,
		"object": OBJECT,
	}
)

type Object map[string]any
type List []any

func DataTypeOf(s string) DataType {
	return dataType[s]
}

func (d DataType) String() string {
	return dataTypeString[dataTypeIndex[d]:dataTypeIndex[d+1]]
}

// ----------------------------------------------------------------------------
// API
// api.endpoint.method.request

type Api struct {
	Protocol  Protocol
	Auth      *Api
	Url       string
	Resources *data.Data

	// the following are global elements
	// that apply to all resources

	Params   *data.Data // url params
	Header   *data.Data // request header
	Request  *data.Data // request body
	Response *data.Data // response body
}

func New(protocol Protocol, url string) *Api {
	return &Api{
		Protocol:  protocol,
		Url:       url,
		Resources: data.Make[*Resource](4),
		Params:    data.Make[*param](4),
		Header:    data.Make[*param](4),
		Request:   data.Make[*param](4),
		Response:  data.Make[*param](4),
	}
}

func FromYaml(yaml []byte) []*Api {
	return nil
}

func FromMap(m map[string]any) (api *Api) {
	if url, ok := m["url"]; ok {
		api = New(REST, url.(string))
		for k, v := range m {
			var d *data.Data
			switch k {
			case "params":
				d = api.Params
			case "header":
				d = api.Header
			case "request":
				d = api.Request
			case "response":
				d = api.Response
			}
			if d != nil {
				switch v := v.(type) {
				case map[string]any:
					_, p := ParamMap(v)
					d.Add(p)
				case []any:
					_, p := ParamList(v)
					d.Add(p)
				}
			}
		}
		// add resources
		// range resources as []any
		// need func that intakes a map[string]any
		// and returns a *Resource
		// add resource to api.Resources
		if r, ok := m["resources"]; ok {
			for k, v := range r.(map[string]any) {
				api.Resources.Add(ResourceFromMap(k, v))
			}
		}
	}
	return
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

// ----------------------------------------------------------------------------
// API RESOURCE

type Resource struct {
	Name      string
	Uri       string
	Methods   *data.Data
	Resources *data.Data
	Params    *data.Data
	Header    *data.Data
	Data      *data.Data
}

func NewResource(name, uri string) *Resource {
	return &Resource{
		Name:      name,
		Uri:       uri,
		Methods:   data.Make[*Method](4),
		Resources: data.Make[*Resource](4),
		Params:    data.Make[*param](4),
		Header:    data.Make[*param](4),
		Data:      data.Make[*param](4),
	}
}

func ResourceMap(k string, v map[string]any) (r *Resource) {
	if uri, ok := v["uri"]; ok {
		r = NewResource(k, uri.(string))
		for k, v := range v {
			var d *data.Data
			switch k {
			case "params":
				d = r.Params
			case "header":
				d = r.Header
			case "data":
				d = r.Data
			}
			if d != nil {
				switch v := v.(type) {
				case map[string]any:
					_, p := ParamMap(v)
					d.Add(p)
				case []any:
					_, p := ParamList(v)
					d.Add(p)
				}
			}
		}
	}
	return
}

func (r *Resource) Key() string {
	return r.Name
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

// ----------------------------------------------------------------------------
// PARAM

// Param is an element of an object or list
// which may be found in the url, header, or
// body of a request or response
type param struct {
	// if the param belongs to an object
	// the key will be the field name,
	// otherwise the param belongs to a list
	// and the key will be the index
	key string
	// the datatype of the param
	typ DataType
	// if the param is an object or aclist with
	// a single datatype and variable length,
	// the elm will be the datatype of the
	// elements in the object or list
	elm DataType
	// if the param is an object or a list
	// the els will be the data elements
	// in the object or list
	els *data.Data
}

func Param(key string, typ, elem *param) *param {
	return &param{
		key: key,
		/* typ: typ,
		elm: elem,
		els: data.Make[any](), */
	}
}

func ParamMap(m map[string]any) (e DataType, d *data.Data) {
	d = data.Make[*param](len(m))
	i := 0
	for k, v := range m {
		p := ParamElem(k, v)
		if i == 0 {
			e = p.typ
		} else if p.typ != e && e != NONE {
			e = NONE
		}
		d.Add(p)
		i++
	}
	return
}

func ParamList(l []any) (e DataType, d *data.Data) {
	d = data.Make[*param](len(l))
	for i, v := range l {
		p := ParamElem(strconv.Itoa(i), v)
		if i == 0 {
			e = p.typ
		} else if p.typ != e && e != NONE {
			e = NONE
		}
		d.Add(p)
	}
	return
}

func ParamElem(k string, v any) (p *param) {
	p = &param{key: k}
	switch v := v.(type) {
	case string:
		p.typ = DataTypeOf(v)
	case map[string]any:
		p.typ = OBJECT
		p.elm, p.els = ParamMap(v)
	case []any:
		p.typ = LIST
		p.elm, p.els = ParamList(v)
	}
	return
}

func (p *param) Key() string {
	return p.key
}
