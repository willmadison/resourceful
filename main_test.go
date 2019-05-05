package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterEndpoints(t *testing.T) {
	h := registerEndpoints()
	handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		h.ServeHTTP(response, request)
	})

	request := httptest.NewRequest(http.MethodPost, "http://example.com/inbox", nil)
	w := httptest.NewRecorder()

	handler(w, request)

	response := w.Result()

	assert.Equal(t, 200, response.StatusCode)
}
