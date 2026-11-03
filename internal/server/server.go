package server

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ironcore-dev/ironcore-dashboard/internal/api"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"nhooyr.io/websocket"
)

type Server struct {
	router   *chi.Mux
	ironcore versioned.Interface
	k8s      *kubernetes.Clientset
	dyn      dynamic.Interface
}

// NewSREOnly serves the local fixture console without constructing or contacting
// any Kubernetes client. End-user API routes are deliberately absent in this mode.
func NewSREOnly(frontendFS fs.FS, sreHandler, metricsHandler http.Handler) *Server {
	s := &Server{}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"}, AllowedMethods: []string{"GET", "OPTIONS"}, AllowedHeaders: []string{"Content-Type"}}))
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	if sreHandler != nil {
		r.Mount("/api/sre/v1", sreHandler)
	}
	if metricsHandler != nil {
		r.Get("/metrics", metricsHandler.ServeHTTP)
	}
	if frontendFS != nil {
		fileServer := http.FileServer(http.FS(frontendFS))
		r.Get("/*", func(w http.ResponseWriter, request *http.Request) {
			if _, err := fs.Stat(frontendFS, request.URL.Path[1:]); err != nil {
				request.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, request)
		})
	}
	s.router = r
	return s
}

func New(cs versioned.Interface, k8sClient *kubernetes.Clientset, dynClient dynamic.Interface, frontendFS fs.FS) *Server {
	return NewWithSRE(cs, k8sClient, dynClient, frontendFS, nil)
}

func NewWithSRE(cs versioned.Interface, k8sClient *kubernetes.Clientset, dynClient dynamic.Interface, frontendFS fs.FS, sreHandler http.Handler) *Server {
	return NewWithSREMetrics(cs, k8sClient, dynClient, frontendFS, sreHandler, nil)
}

func NewWithSREMetrics(cs versioned.Interface, k8sClient *kubernetes.Clientset, dynClient dynamic.Interface, frontendFS fs.FS, sreHandler, metricsHandler http.Handler) *Server {
	s := &Server{ironcore: cs, k8s: k8sClient, dyn: dynClient}
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

	// Servers (metal-operator)
	sh := api.NewServerHandler(s.dyn)
	r.Get("/api/v1/servers", sh.List)

	// Machine WebSocket console
	r.Get("/api/v1/ws/machines/{ns}/{name}/console", s.handleMachineConsole)

	if sreHandler != nil {
		r.Mount("/api/sre/v1", sreHandler)
	}
	if metricsHandler != nil {
		r.Get("/metrics", metricsHandler.ServeHTTP)
	}

	// Serve built Vue SPA for all other routes
	if frontendFS != nil {
		fileServer := http.FileServer(http.FS(frontendFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// SPA fallback: serve index.html for unknown routes so vue-router handles them
			_, statErr := fs.Stat(frontendFS, r.URL.Path[1:])
			if statErr != nil {
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	s.router = r
	return s
}

func (s *Server) handleMachineConsole(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	name := chi.URLParam(r, "name")

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	banner := "\r\n\033[1;32mIronCore Machine Console — " + name + " (" + ns + ")\033[0m\r\n\033[90mConnecting...\033[0m\r\n"
	conn.Write(ctx, websocket.MessageText, []byte(banner))

	// Try to get machine IPs via NIC objects and SSH passthrough
	machine, err := s.ironcore.ComputeV1alpha1().Machines(ns).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		for _, niStatus := range machine.Status.NetworkInterfaces {
			if niStatus.NetworkInterfaceRef.Name == "" {
				continue
			}
			nic, nicErr := s.ironcore.NetworkingV1alpha1().NetworkInterfaces(ns).Get(ctx, niStatus.NetworkInterfaceRef.Name, metav1.GetOptions{})
			if nicErr != nil {
				continue
			}
			for _, ipAddr := range nic.Status.IPs {
				ip := ipAddr.String()
				tcpConn, dialErr := net.DialTimeout("tcp", ip+":22", 3_000_000_000)
				if dialErr == nil {
					conn.Write(ctx, websocket.MessageText, []byte("\033[32mSSH passthrough to "+ip+":22\033[0m\r\n"))
					go func() {
						for {
							_, msg, readErr := conn.Read(ctx)
							if readErr != nil {
								return
							}
							tcpConn.Write(msg)
						}
					}()
					io.Copy(&wsWriter{conn: conn, ctx: ctx}, tcpConn)
					return
				}
			}
		}
	}

	// Simulated shell fallback
	conn.Write(ctx, websocket.MessageText, []byte("\033[33mNo reachable console endpoint. Running simulated shell.\033[0m\r\n\r\n\033[1m"+name+":~$\033[0m "))
	for {
		_, msg, readErr := conn.Read(ctx)
		if readErr != nil {
			return
		}

		var frame struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(msg, &frame) == nil && frame.Type == "resize" {
			continue
		}

		line := string(msg)
		if line == "exit\r" || line == "exit\n" || line == "exit" {
			conn.Write(ctx, websocket.MessageText, []byte("\r\nlogout\r\n"))
			conn.Close(websocket.StatusNormalClosure, "logout")
			return
		}
		response := "\r\n\033[33mbash: " + line + ": command not found\033[0m\r\n\033[1m" + name + ":~$\033[0m "
		conn.Write(ctx, websocket.MessageText, []byte(response))
	}
}

type wsWriter struct {
	conn *websocket.Conn
	ctx  context.Context
}

func (w *wsWriter) Write(p []byte) (int, error) {
	err := w.conn.Write(w.ctx, websocket.MessageBinary, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
