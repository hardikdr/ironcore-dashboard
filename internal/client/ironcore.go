package client

import (
	"fmt"

	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// New returns an IronCore versioned clientset.
// If kubeconfigPath is empty it falls back to in-cluster config.
func New(kubeconfigPath string) (versioned.Interface, error) {
	var cfg *rest.Config
	var err error

	if kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig: %w", err)
	}

	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build ironcore clientset: %w", err)
	}
	return cs, nil
}
