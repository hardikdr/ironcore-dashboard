package server

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ironcore-dev/ironcore-dashboard/internal/api"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	router   *chi.Mux
	ironcore versioned.Interface
	k8s      *kubernetes.Clientset
}

func New(cs versioned.Interface, k8sClient *kubernetes.Clientset, frontendFS fs.FS) *Server {
	s := &Server{ironcore: cs, k8s: k8sClient}
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Namespace list (for project/namespace switcher)
	r.Get("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
		list, err := s.k8s.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		names := make([]string, 0, len(list.Items))
		for _, ns := range list.Items {
			names = append(names, ns.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(names)
	})

	// Machine routes
	mh := api.NewMachineHandler(cs)
	r.Get("/api/v1/machineclasses", mh.ListMachineClasses)
	r.Route("/api/v1/namespaces/{ns}/machines", func(r chi.Router) {
		r.Get("/", mh.List)
		r.Post("/", mh.Create)
		r.Get("/{name}", mh.Get)
		r.Delete("/{name}", mh.Delete)
		r.Patch("/{name}/power", mh.PatchPower)
	})

	// Volume routes
	vh := api.NewVolumeHandler(cs)
	vch := api.NewVolumeClassHandler(cs)
	r.Get("/api/v1/volumeclasses", vch.List)
	r.Route("/api/v1/namespaces/{ns}/volumes", func(r chi.Router) {
		r.Get("/", vh.List)
		r.Post("/", vh.Create)
		r.Get("/{name}", vh.Get)
		r.Delete("/{name}", vh.Delete)
	})

	// Networking routes
	nh := api.NewNetworkHandler(cs)
	vip := api.NewVirtualIPHandler(cs)
	lb := api.NewLoadBalancerHandler(cs)
	r.Route("/api/v1/namespaces/{ns}/networks", func(r chi.Router) {
		r.Get("/", nh.ListNetworks)
		r.Post("/", nh.CreateNetwork)
		r.Get("/{name}", nh.GetNetwork)
		r.Delete("/{name}", nh.DeleteNetwork)
	})
	r.Get("/api/v1/namespaces/{ns}/networkinterfaces", nh.ListNetworkInterfaces)
	r.Route("/api/v1/namespaces/{ns}/virtualips", func(r chi.Router) {
		r.Get("/", vip.List)
		r.Post("/", vip.Create)
		r.Get("/{name}", vip.Get)
		r.Delete("/{name}", vip.Delete)
	})
	r.Route("/api/v1/namespaces/{ns}/loadbalancers", func(r chi.Router) {
		r.Get("/", lb.List)
		r.Post("/", lb.Create)
		r.Get("/{name}", lb.Get)
		r.Delete("/{name}", lb.Delete)
	})

	// IPAM routes
	iph := api.NewIPAMHandler(cs)
	r.Route("/api/v1/namespaces/{ns}/prefixes", func(r chi.Router) {
		r.Get("/", iph.ListPrefixes)
		r.Post("/", iph.CreatePrefix)
		r.Get("/{name}", iph.GetPrefix)
		r.Delete("/{name}", iph.DeletePrefix)
	})

	// Serve built Vue SPA for all other routes
	if frontendFS != nil {
		fileServer := http.FileServer(http.FS(frontendFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// SPA fallback: serve index.html for unknown routes so vue-router handles them
			_, statErr := frontendFS.(fs.StatFS).Stat(r.URL.Path[1:])
			if statErr != nil {
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	s.router = r
	return s
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
