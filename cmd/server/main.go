package main

import (
	"flag"
	"log"

	ironclient "github.com/ironcore-dev/ironcore-dashboard/internal/client"
	"github.com/ironcore-dev/ironcore-dashboard/internal/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	addr       := flag.String("addr", ":8080", "HTTP listen address")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (empty = in-cluster)")
	flag.Parse()

	cs, err := ironclient.New(*kubeconfig)
	if err != nil {
		log.Fatalf("ironcore client: %v", err)
	}

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

	srv := server.New(cs, k8sClient)
	log.Printf("IronCore Dashboard listening on %s", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
