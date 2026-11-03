package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/observability"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

const MaxResponseBytes = 2 << 20

type API struct {
	store      *snapshot.Store
	staleAfter time.Duration
	metrics    *observability.Metrics
	etagSeed   string
}

func New(store *snapshot.Store, staleAfter time.Duration) *API {
	return NewWithMetrics(store, staleAfter, observability.NewMetrics())
}

func NewWithMetrics(store *snapshot.Store, staleAfter time.Duration, metrics *observability.Metrics) *API {
	if staleAfter <= 0 {
		staleAfter = 2 * time.Minute
	}
	if metrics == nil {
		metrics = observability.NewMetrics()
	}
	seed := make([]byte, 8)
	if _, err := rand.Read(seed); err != nil {
		seed = []byte(fmt.Sprintf("%016x", time.Now().UnixNano()))
	}
	return &API{store: store, staleAfter: staleAfter, metrics: metrics, etagSeed: hex.EncodeToString(seed)}
}

func (api *API) Router() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Timeout(10 * time.Second))
	router.Use(api.observeRequests)
	router.Get("/status", api.status)
	router.Get("/sources", api.sources)
	router.Get("/summary", api.summary)
	router.Get("/search", api.search)
	router.Get("/zones", api.zones)
	router.Get("/zones/{zone}", api.zone)
	router.Get("/nodes/{id}", api.node)
	router.Get("/vms", api.vms)
	router.Get("/vms/{id}", api.vm)
	router.Get("/shoots", api.shoots)
	router.Get("/shoots/{id}", api.shoot)
	router.Get("/networks/{id}", api.network)
	router.Get("/loadbalancers/{id}", api.loadBalancer)
	router.Get("/volumes/{id}", api.volume)
	router.Get("/observations", api.observations)
	router.Get("/graph", api.graph)
	router.Get("/resources/{id}/sanitized", api.sanitized)
	return router
}

type metricWriter struct {
	http.ResponseWriter
	status  int
	partial bool
}

func (writer *metricWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}
func (writer *metricWriter) Write(body []byte) (int, error) {
	if writer.status == 0 {
		writer.status = http.StatusOK
	}
	return writer.ResponseWriter.Write(body)
}
func (writer *metricWriter) markPartial() { writer.partial = true }

func (api *API) observeRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		started := time.Now()
		writer := &metricWriter{ResponseWriter: w}
		next.ServeHTTP(writer, request)
		status := writer.status
		if status == 0 {
			status = http.StatusOK
		}
		route := chi.RouteContext(request.Context()).RoutePattern()
		if route == "" {
			route = "/unknown"
		}
		api.metrics.ObserveAPI(route, status, time.Since(started), writer.partial)
	})
}

func EncodeID(key model.ResourceKey) string {
	return base64.RawURLEncoding.EncodeToString([]byte(key))
}

func DecodeID(id string) (model.ResourceKey, error) {
	if id == "" || len(id) > 4096 {
		return "", fmt.Errorf("invalid resource id")
	}
	raw, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return "", fmt.Errorf("invalid resource id")
	}
	key := model.ResourceKey(raw)
	if _, err := model.ParseResourceKey(key); err != nil {
		return "", fmt.Errorf("invalid resource id")
	}
	return key, nil
}
