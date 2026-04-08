package server

import (
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

func New(cs versioned.Interface, k8sClient *kubernetes.Clientset) *Server {
	s := &Server{ironcore: cs, k8s: k8sClient}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Namespace list
	r.Get("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
		list, err := s.k8s.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		names := make([]string, 0, len(list.Items))
		for _, ns := range list.Items {
			names = append(names, ns.Name)
		}
		api.WriteJSON(w, http.StatusOK, names)
	})

	mh := api.NewMachineHandler(cs)
	r.Get("/api/v1/machineclasses", mh.ListMachineClasses)
	r.Route("/api/v1/namespaces/{ns}/machines", func(r chi.Router) {
		r.Get("/", mh.List)
		r.Post("/", mh.Create)
		r.Get("/{name}", mh.Get)
		r.Delete("/{name}", mh.Delete)
		r.Patch("/{name}/power", mh.PatchPower)
	})

	vh := api.NewVolumeHandler(cs)
	nh := api.NewNetworkHandler(cs)
	vip := api.NewVirtualIPHandler(cs)
	lb := api.NewLoadBalancerHandler(cs)
	iph := api.NewIPAMHandler(cs)

	r.Route("/api/v1/namespaces/{ns}", func(r chi.Router) {
		r.Get("/volumes", vh.List)
		r.Post("/volumes", vh.Create)
		r.Delete("/volumes/{name}", vh.Delete)
		r.Get("/networks", nh.ListNetworks)
		r.Get("/networkinterfaces", nh.ListNetworkInterfaces)
		r.Get("/virtualips", vip.List)
		r.Get("/loadbalancers", lb.List)
		r.Get("/prefixes", iph.ListPrefixes)
	})

	s.router = r
	return s
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
