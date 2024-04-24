package api

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/jcdotter/go/test"
)

func TestApi(t *testing.T) {
	path := "https://api.sampleapis.com/csscolornames/colors"
	r, err := http.Get(path + "/1")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if r.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, r.StatusCode)
	}
	fmt.Println(r.Status)
	b := make([]byte, 100)
	r.Body.Read(b)
	fmt.Println(string(b))
	r.Body.Close()
}

func TestDataType(t *testing.T) {
	test.Assert(t, INT.String(), "int")
	test.Assert(t, STRING.String(), "string")
	test.Assert(t, FLOAT.String(), "float")
	test.Assert(t, BOOL.String(), "bool")
	test.Assert(t, LIST.String(), "list")
	test.Assert(t, OBJECT.String(), "object")
}

func TestUrl(t *testing.T) {
	u := "https://api.sampleapis.com/"
	r, _ := url.Parse(u)
	fmt.Println(r.Scheme)
	fmt.Println(r.Host)
	r.Path = "/csscolornames/colors/:id"
	fmt.Println(r.String())
}

func TestParam(t *testing.T) {
	l := map[string]any{
		"header": map[string]any{
			"content-type": "application/json",
		},
		"body": []any{
			map[string]any{
				"id":   "int",
				"name": "string",
				"hex":  "string",
			},
		},
	}
	dt, d := ParamMap(l)
	test.Assert(t, dt, ANY, "expected param elem type of ANY, got %v", dt)

	e := d.Get("header")
	test.Assert(t, e.Key(), "header", "expected param key 'header' got '%v'", e.Key())
	test.Assert(t, e.Type(), OBJECT, "expected param type OBJECT got %v", e.Type())
	test.Assert(t, e.ElemType(), STRING, "expected param elem type of STRING, got %v", e.ElemType())

	e = e.Elems().Get("content-type")
	test.Assert(t, e.Key(), "content-type", "expected param key 'content-type' got '%v'", e.Key())
	test.Assert(t, e.Type(), STRING, "expected param type STRING got %v", e.Type())
	test.Assert(t, e.ElemType(), NONE, "expected param elem type of NONE, got %v", e.ElemType())
	test.Assert(t, e.Elems().IsNil(), true, "expected param elems is nil, got %v", e.Elems().IsNil())

	e = d.Get("body")
	test.Assert(t, e.Key(), "body", "expected param key 'body' got '%v'", e.Key())
	test.Assert(t, e.Type(), LIST, "expected param type LIST got %v", e.Type())
	test.Assert(t, e.ElemType(), OBJECT, "expected param elem type of OBJECT, got %v", e.ElemType())

	e = e.Elems().Index(0)
	test.Assert(t, e.Type(), OBJECT, "expected param type OBJECT got %v", e.Type())
	test.Assert(t, e.ElemType(), ANY, "expected param elem type of ANY, got %v", e.ElemType())

	x := e.Elems().Get("id")
	test.Assert(t, x.Type(), INT, "expected param type INT got %v", x.Type())

	x = e.Elems().Get("name")
	test.Assert(t, x.Type(), STRING, "expected param type STRING got %v", x.Type())

	x = e.Elems().Get("hex")
	test.Assert(t, x.Type(), STRING, "expected param type STRING got %v", x.Type())
}

func TestYaml(t *testing.T) {
	yml, _ := os.ReadFile("api.yml")
	api := FromYaml(yml)

	test.Assert(t, api.Protocol, REST, "expected protocol REST, got %v", api.Protocol)
	test.Assert(t, api.Resources.Len(), 2, "expected 2 resources, got %v", api.Resources.Len())

	typ := api.Resource("color").Method("GET").Response.Header.Get("content-type").Type()
	test.Assert(t, typ, STRING, "expected content-type string, got %v", typ)

	typ = api.Resource("colors").Method("GET").Response.Body.Index(0).Elem("id").Type()
	test.Assert(t, typ, INT, "expected id as int, got %v", typ)
}

func TestCall(t *testing.T) {
	/* yml, _ := os.ReadFile("api.yml")
	api := FromYaml(yml)

	r, err := api.Resource("color").Method("GET").Call()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if r.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, r.StatusCode)
	}
	fmt.Println(r.Status)
	b := make([]byte, 100)
	r.Body.Read(b)
	fmt.Println(string(b))
	r.Body.Close() */
}
