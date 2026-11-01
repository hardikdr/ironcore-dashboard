package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDemoServer_HealthZ(t *testing.T) {
	s := NewDemo(nil)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthz", nil)
	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDemoServer_ModeEndpoint(t *testing.T) {
	s := NewDemo(nil)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/mode", nil)
	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != `{"demo":true}` {
		t.Errorf("unexpected mode body: %s", body)
	}
}

func TestDemoServer_Namespaces(t *testing.T) {
	s := NewDemo(nil)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/namespaces", nil)
	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDemoServer_MachineList(t *testing.T) {
	s := NewDemo(nil)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/namespaces/default/machines", nil)
	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDemoServer_UnknownRoute404(t *testing.T) {
	s := NewDemo(nil)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nonexistent/resource", nil)
	s.Router().ServeHTTP(rec, req)

	// demo handler returns 404 for unknown resources
	if rec.Code == http.StatusInternalServerError {
		t.Fatal("should not 500 on unknown routes")
	}
}
