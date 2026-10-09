// Starter test for env() so `go test ./...` does real work in CI. No DB needed.
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEnvReturnsFallbackWhenUnset(t *testing.T) {
	os.Unsetenv("DEVBOARD_TEST_KEY")
	if got := env("DEVBOARD_TEST_KEY", "fallback"); got != "fallback" {
		t.Errorf("env() = %q, want %q", got, "fallback")
	}
}

func TestEnvReturnsValueWhenSet(t *testing.T) {
	os.Setenv("DEVBOARD_TEST_KEY", "real")
	defer os.Unsetenv("DEVBOARD_TEST_KEY")
	if got := env("DEVBOARD_TEST_KEY", "fallback"); got != "real" {
		t.Errorf("env() = %q, want %q", got, "real")
	}
}

func TestServeFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	frontendDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(frontendDir, "index.html"),
		[]byte("<main>DevBoard</main>"),
		0600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(frontendDir, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(frontendDir, "assets", "app.js"),
		[]byte("console.log('ready')"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.NoRoute(serveFrontend(frontendDir))
	testCases := []struct {
		method     string
		path       string
		statusCode int
		body       string
	}{
		{method: http.MethodGet, path: "/", statusCode: http.StatusOK, body: "<main>DevBoard</main>"},
		{method: http.MethodGet, path: "/projects/1", statusCode: http.StatusOK, body: "<main>DevBoard</main>"},
		{method: http.MethodGet, path: "/assets/app.js", statusCode: http.StatusOK, body: "console.log('ready')"},
		{method: http.MethodGet, path: "/assets/missing.js", statusCode: http.StatusNotFound, body: "not found"},
		{method: http.MethodGet, path: "/api/missing", statusCode: http.StatusNotFound, body: "not found"},
		{method: http.MethodPost, path: "/unknown", statusCode: http.StatusNotFound, body: "not found"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.method+" "+testCase.path, func(t *testing.T) {
			request := httptest.NewRequest(testCase.method, testCase.path, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != testCase.statusCode {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.statusCode)
			}
			if !strings.Contains(recorder.Body.String(), testCase.body) {
				t.Fatalf("body %q does not contain %q", recorder.Body.String(), testCase.body)
			}
		})
	}
}
