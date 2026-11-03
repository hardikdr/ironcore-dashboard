package server

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMountingSRERouterDoesNotChangeUserAPIContract(t *testing.T) {
	t.Parallel()

	baseline := New(nil, nil, nil, nil)
	sre := chi.NewRouter()
	sre.Get("/status", func(http.ResponseWriter, *http.Request) {})
	withSRE := NewWithSRE(nil, nil, nil, nil, sre)

	baselineRoutes := userRoutes(t, baseline.Router())
	withSRERoutes := userRoutes(t, withSRE.Router())
	if strings.Join(baselineRoutes, "\n") != strings.Join(withSRERoutes, "\n") {
		t.Fatalf("/api/v1 contract changed\nbaseline: %#v\nwith SRE: %#v", baselineRoutes, withSRERoutes)
	}
}

func userRoutes(t *testing.T, router *chi.Mux) []string {
	t.Helper()
	var routes []string
	if err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if strings.HasPrefix(route, "/api/v1") {
			routes = append(routes, method+" "+route)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(routes)
	return routes
}
