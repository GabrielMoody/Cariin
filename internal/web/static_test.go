package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesSearchPage(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	Handler().ServeHTTP(recorder, request)

	if recorder.Code != 200 {
		t.Fatalf("Handler() status = %d, want 200", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Handler() content type = %q, want text/html", contentType)
	}
	if !strings.Contains(recorder.Body.String(), `id="search-form"`) {
		t.Fatal("Handler() did not serve the search form")
	}
}
