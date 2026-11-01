package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newDemoReq(method, path string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)
	return req
}

func TestDemoHandler_Namespaces(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, newDemoReq("GET", "/namespaces"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected at least one namespace")
	}
}

func TestDemoHandler_MachineList(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, newDemoReq("GET", "/namespaces/default/machines"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []MachineResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected machines in demo fixture")
	}
}

func TestDemoHandler_MachineGet_AllFields(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, newDemoReq("GET", "/namespaces/default/machines/web-frontend-01"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var m MachineResponse
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Verify all fields are populated (this caught the VolumeDetailResponse bug)
	if m.Name == "" {
		t.Error("Name is empty")
	}
	if m.MachineClass == "" {
		t.Error("MachineClass is empty — detail view would show blank")
	}
	if m.Image == "" {
		t.Error("Image is empty — detail view would show blank")
	}
	if m.State == "" {
		t.Error("State is empty — detail view would show blank")
	}
	if m.Power == "" {
		t.Error("Power is empty — detail view would show blank")
	}
}

func TestDemoHandler_MachineGet_NotFound(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, newDemoReq("GET", "/namespaces/default/machines/no-such-machine"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDemoHandler_VolumeGet_DetailShape(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, newDemoReq("GET", "/namespaces/default/volumes/data-01"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var v VolumeDetailResponse
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if v.VolumeID == "" {
		t.Error("VolumeID is empty")
	}
	if v.SizeGiB == 0 {
		t.Error("SizeGiB is 0 — bytes->GiB conversion failed")
	}
}

func TestDemoHandler_EventsFiltered(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	req := newDemoReq("GET", "/namespaces/default/events?regarding=web-frontend-01")
	d.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []EventResponse
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, ev := range events {
		if ev.Regarding != "web-frontend-01" {
			t.Errorf("got event regarding %q, expected web-frontend-01", ev.Regarding)
		}
	}
}

func TestDemoHandler_MachineCreate(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	body := `{"name":"test-machine","machineClass":"x1-small","image":"gardenlinux@latest"}`
	req, _ := http.NewRequest("POST", "/namespaces/default/machines", strings.NewReader(body))
	d.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestDemoHandler_MachineDelete(t *testing.T) {
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/namespaces/default/machines/web-frontend-01", nil)
	d.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDemoHandler_Mode_NotHandled(t *testing.T) {
	// /mode is handled by server.go before DemoHandler is mounted — demo handler should 404
	d := &DemoHandler{}
	rec := httptest.NewRecorder()
	d.ServeHTTP(rec, newDemoReq("GET", "/mode"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 from demo handler for /mode, got %d", rec.Code)
	}
}
