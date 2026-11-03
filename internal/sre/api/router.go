package api

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

const MaxResponseBytes = 2 << 20

type API struct {
	store      *snapshot.Store
	staleAfter time.Duration
}

func New(store *snapshot.Store, staleAfter time.Duration) *API {
	if staleAfter <= 0 {
		staleAfter = 2 * time.Minute
	}
	return &API{store: store, staleAfter: staleAfter}
}

func (api *API) Router() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Timeout(10 * time.Second))
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
