package api

import (
	"fmt"
	"net/http"
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

func TestNew(t *testing.T) {
	/* a := &Api{
		Protocol: REST,
		Url:      "https://api.sampleapis.com/",
		Resources: data.Of(
			&Resource{
				Uri: "csscolornames/colors",
				Methods: data.Of(
					&Method{
						Name: "GET",
						Response: &Response{
							Header: data.Of(
								Param("content-type", "application/json"),
							),
							Body: List{
								Object{
									"id":   INT,
									"name": STRING,
									"hex":  STRING,
								},
							},
						},
					},
				),
			},
		),
	}

	// []any, map[string]any, BOOL, INT, FLOAT, STRING
	_ = List{
		Object{
			"id":   INT,
			"name": STRING,
			"hex":  STRING,
		},
	} */
}

func TestDataType(t *testing.T) {

	test.Assert(t, INT.String(), "int")
	test.Assert(t, STRING.String(), "string")
	test.Assert(t, FLOAT.String(), "float")
	test.Assert(t, BOOL.String(), "bool")
	test.Assert(t, LIST.String(), "list")
	test.Assert(t, OBJECT.String(), "object")

}

func TestParam(t *testing.T) {

}
