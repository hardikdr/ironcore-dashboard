package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"time"

	rootpkg "github.com/ironcore-dev/ironcore-dashboard"
	ironclient "github.com/ironcore-dev/ironcore-dashboard/internal/client"
	"github.com/ironcore-dev/ironcore-dashboard/internal/server"
	sreapi "github.com/ironcore-dev/ironcore-dashboard/internal/sre/api"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/collector"
	sreconfig "github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	srefixture "github.com/ironcore-dev/ironcore-dashboard/internal/sre/fixture"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/observability"
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
	sreFixtureDir := flag.String("sre-fixture-dir", "", "path to sanitized SRE JSON fixtures for local-only mode (mutually exclusive with --sre-config)")
	flag.Parse()
	if *sreConfig != "" && *sreFixtureDir != "" {
		log.Fatal("--sre-config and --sre-fixture-dir are mutually exclusive")
	}

	var sreHandler http.Handler
	var metricsHandler http.Handler
	if *sreFixtureDir != "" {
		store, err := srefixture.LoadDir(*sreFixtureDir)
		if err != nil {
			log.Fatalf("SRE fixtures: %v", err)
		}
		metrics := observability.NewMetrics()
		sreHandler = sreapi.NewWithMetrics(store, 2*time.Minute, metrics).Router()
		metricsHandler = metrics.Handler()
		log.Printf("SRE Console fixture mode enabled from sanitized local snapshots")
		srv := server.NewSREOnly(rootpkg.Frontend(), sreHandler, metricsHandler)
		log.Printf("IronCore Dashboard local fixture server listening on %s", *addr)
		if err := srv.ListenAndServe(*addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
		return
	}
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
			metrics := observability.NewMetrics()
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
			scheduler := collector.NewScheduler(collectors, store, collector.SchedulerOptions{MaxConcurrent: cfg.MaxConcurrent, Interval: cfg.PollInterval, JitterFraction: 0.1, Metrics: metrics, Logger: slog.Default()})
			go func() {
				if err := scheduler.Run(context.Background()); err != nil && err != context.Canceled {
					log.Printf("SRE collector scheduler stopped: %v", err)
				}
			}()
			sreHandler = sreapi.NewWithMetrics(store, cfg.StaleAfter, metrics).Router()
			metricsHandler = metrics.Handler()
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

	srv := server.NewWithSREMetrics(cs, k8sClient, dynClient, rootpkg.Frontend(), sreHandler, metricsHandler)
	log.Printf("IronCore Dashboard listening on %s", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
