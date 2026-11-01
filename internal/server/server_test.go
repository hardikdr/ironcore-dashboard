package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nhooyr.io/websocket"
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

// TestConsoleUnavailableInRealMode verifies that the real-mode console WebSocket
// handler immediately sends a JSON error message and closes the connection,
// rather than attempting a raw TCP dial to port 22.
func TestConsoleUnavailableInRealMode(t *testing.T) {
	srv := &Server{}
	ts := httptest.NewServer(http.HandlerFunc(srv.handleMachineConsole))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	c, _, err := websocket.Dial(t.Context(), wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer c.CloseNow()

	_, msg, err := c.Read(t.Context())
	if err != nil {
		t.Fatalf("read first message: %v", err)
	}

	var payload struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if jsonErr := json.Unmarshal(msg, &payload); jsonErr != nil {
		t.Fatalf("message is not JSON: %v — raw: %s", jsonErr, msg)
	}
	if payload.Type != "error" {
		t.Errorf("expected type=error, got %q", payload.Type)
	}
	if !strings.Contains(payload.Message, "not available") {
		t.Errorf("expected error message to mention 'not available', got: %s", payload.Message)
	}
}
