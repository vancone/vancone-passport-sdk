package util

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"method":"` + r.Method + `","body":` + string(body) + `}`))
	}))
	defer server.Close()

	resp := Request(server.URL, "POST", `{"a":1}`, false)
	if !strings.Contains(string(resp), `"method":"POST"`) || !strings.Contains(string(resp), `"body":{"a":1}`) {
		t.Fatalf("unexpected response: %s", resp)
	}
}
