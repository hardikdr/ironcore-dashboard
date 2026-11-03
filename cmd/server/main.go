package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	rootpkg "github.com/ironcore-dev/ironcore-dashboard"
	ironclient "github.com/ironcore-dev/ironcore-dashboard/internal/client"
	"github.com/ironcore-dev/ironcore-dashboard/internal/server"
	sreapi "github.com/ironcore-dev/ironcore-dashboard/internal/sre/api"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/collector"
	sreconfig "github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/source"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (empty = in-cluster)")
	sreConfig := flag.String("sre-config", "", "path to read-only multi-source SRE configuration (empty = disabled)")
	flag.Parse()

	var sreHandler http.Handler
	if *sreConfig != "" {
		cfg, err := sreconfig.Load(*sreConfig)
		if err != nil {
			log.Fatalf("SRE configuration: %v", err)
		}
		registry, err := source.NewRegistry(context.Background(), cfg)
		if err != nil {
			log.Fatalf("SRE source registry: %v", err)
		}
		if cfg.Enabled {
			store := snapshot.NewStore()
			collectors := make([]collector.Collector, 0, registry.Count())
			for _, entry := range registry.Entries() {
				dynamicClient, err := dynamic.NewForConfig(entry.Config)
				if err != nil {
					log.Fatalf("SRE source %q dynamic client: %v", entry.Metadata.ID, err)
				}
				discoveryClient, err := discovery.NewDiscoveryClientForConfig(entry.Config)
				if err != nil {
					log.Fatalf("SRE source %q discovery client: %v", entry.Metadata.ID, err)
				}
				reader := collector.ClientReader{Dynamic: dynamicClient, Discovery: discoveryClient}
				switch entry.Metadata.Role {
				case sreconfig.RoleZonalCAPI:
					collectors = append(collectors, collector.NewZonalCollector(entry.Metadata, reader, reader, collector.ZonalOptions{RequestTimeout: cfg.RequestTimeout}))
				case sreconfig.RoleCentralIronCore:
					collectors = append(collectors, collector.NewIronCoreCollector(entry.Metadata, reader, collector.IronCoreOptions{RequestTimeout: cfg.RequestTimeout}))
				case sreconfig.RoleGardener:
					collectors = append(collectors, collector.NewGardenerCollector(entry.Metadata, reader, reader, collector.GardenerOptions{RequestTimeout: cfg.RequestTimeout}))
				}
				store.PublishFailure(entry.Metadata.ID, time.Now().UTC(), 0, snapshot.ErrorNone, model.AccessUnknown)
			}
			scheduler := collector.NewScheduler(collectors, store, collector.SchedulerOptions{MaxConcurrent: cfg.MaxConcurrent, Interval: cfg.PollInterval, JitterFraction: 0.1})
			go func() {
				if err := scheduler.Run(context.Background()); err != nil && err != context.Canceled {
					log.Printf("SRE collector scheduler stopped: %v", err)
				}
			}()
			sreHandler = sreapi.New(store, cfg.StaleAfter).Router()
			log.Printf("SRE Console enabled with %d read-only sources", registry.Count())
		}
	}

	// IronCore clientset
	cs, err := ironclient.New(*kubeconfig)
	if err != nil {
		log.Fatalf("ironcore client: %v", err)
	}

	// Standard k8s clientset (for namespace listing)
	var k8sCfg *rest.Config
	if *kubeconfig != "" {
		k8sCfg, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		k8sCfg, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatalf("k8s config: %v", err)
	}
	k8sClient, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		log.Fatalf("k8s client: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(k8sCfg)
	if err != nil {
		log.Fatalf("dynamic client: %v", err)
	}

	srv := server.NewWithSRE(cs, k8sClient, dynClient, rootpkg.Frontend(), sreHandler)
	log.Printf("IronCore Dashboard listening on %s", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
