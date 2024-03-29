package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jcdotter/go/data"
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
	a := &Api{
		Protocol: REST,
		Url:      "https://api.sampleapis.com/",
		Endpoints: Endpoints{data.Of(
			&Endpoint{
				Uri:     "/csscolornames/colors",
				Content: JSON,
				Allow:   JSON,
				Methods: Methods{data.Of(
					&Method{
						Type: GET,
						Request: &Request{
							Path: data.Of(
								Param("id", "int"),
							),
						},
						Response: &Response{},
					},
				)},
			},
		)},
	}
	a.Endpoints.Get("/csscolornames/colors").Methods.Get(GET).Request.Path.Get("id")
}
