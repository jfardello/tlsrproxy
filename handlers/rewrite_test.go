package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "/_tlsrproxy/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	env := Env{}
	handler := http.Handler(Handler{&env, Status})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if rr.Body.String() == "OK\n" {
		t.Log("got a good response.")
	} else {
		t.Fatalf("bad response!")
	}
}
