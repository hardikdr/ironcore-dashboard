package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// DemoHandler serves static fixture data for demo/preview mode.
// Activated by passing a nil IronCore client or DEMO_MODE env var.
// All responses are clearly synthetic — never mask real API failures.
type DemoHandler struct{}

var (
	demoNow   = time.Now().UTC().Format(time.RFC3339)
	demoNow1h = time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	demoNow2h = time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	demoNow1d = time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	demoNow2d = time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
)

var demoFixtures = map[string]any{
	"namespaces": []string{"default", "dev", "staging", "production"},

	"machineclasses": []MachineClassResponse{
		{Name: "x1-small", CPU: "2", RAM: "4Gi"},
		{Name: "x1-medium", CPU: "4", RAM: "8Gi"},
		{Name: "x1-large", CPU: "8", RAM: "16Gi"},
		{Name: "x1-xlarge", CPU: "16", RAM: "32Gi"},
		{Name: "gpu-v100-1", CPU: "8", RAM: "64Gi"},
	},
	"volumeclasses": []VolumeClassResponse{
		{Name: "ssd-standard", Storage: "100Gi"},
		{Name: "ssd-fast", Storage: "500Gi"},
		{Name: "hdd-archive", Storage: "2Ti"},
	},
	"machinepools": []MachinePoolResponse{
		{Name: "pool-eu-central-1a", State: "Ready", AvailableMachineClasses: []string{"x1-small", "x1-medium", "x1-large"}, CreatedAt: demoNow2d},
		{Name: "pool-eu-central-1b", State: "Ready", AvailableMachineClasses: []string{"x1-medium", "x1-xlarge", "gpu-v100-1"}, CreatedAt: demoNow2d},
	},
	"volumepools": []VolumePoolResponse{
		{Name: "vpool-ssd-zone-a", State: "Available", AvailableVolumeClasses: []string{"ssd-standard", "ssd-fast"}, CreatedAt: demoNow2d},
		{Name: "vpool-hdd-zone-a", State: "Available", AvailableVolumeClasses: []string{"hdd-archive"}, CreatedAt: demoNow2d},
	},

	"machines": []MachineResponse{
		{Name: "web-frontend-01", Namespace: "default", State: "Running", Power: "On", MachineClass: "x1-medium", Image: "gardenlinux@sha256:abc123", IPs: []string{"10.0.0.10"}, Volumes: []string{"data-01"}, CreatedAt: demoNow2d},
		{Name: "web-frontend-02", Namespace: "default", State: "Running", Power: "On", MachineClass: "x1-medium", Image: "gardenlinux@sha256:abc123", IPs: []string{"10.0.0.11"}, Volumes: []string{"data-02"}, CreatedAt: demoNow2d},
		{Name: "api-server-01", Namespace: "default", State: "Running", Power: "On", MachineClass: "x1-large", Image: "gardenlinux@sha256:def456", IPs: []string{"10.0.0.20"}, Volumes: []string{"api-data"}, CreatedAt: demoNow1d},
		{Name: "worker-batch-01", Namespace: "default", State: "Pending", Power: "Off", MachineClass: "x1-xlarge", Image: "gardenlinux@sha256:ghi789", IPs: []string{}, Volumes: []string{}, CreatedAt: demoNow1h},
		{Name: "gpu-trainer-01", Namespace: "default", State: "Shutdown", Power: "Off", MachineClass: "gpu-v100-1", Image: "gardenlinux@sha256:jkl012", IPs: []string{"10.0.0.30"}, Volumes: []string{"model-store"}, CreatedAt: demoNow2h},
	},

	"volumes": []VolumeResponse{
		{Name: "data-01", Namespace: "default", State: "Available", SizeBytes: 100 * 1024 * 1024 * 1024, VolumeClass: "ssd-standard", CreatedAt: demoNow2d},
		{Name: "data-02", Namespace: "default", State: "Available", SizeBytes: 100 * 1024 * 1024 * 1024, VolumeClass: "ssd-standard", CreatedAt: demoNow2d},
		{Name: "api-data", Namespace: "default", State: "Available", SizeBytes: 500 * 1024 * 1024 * 1024, VolumeClass: "ssd-fast", CreatedAt: demoNow1d},
		{Name: "model-store", Namespace: "default", State: "Available", SizeBytes: 2 * 1024 * 1024 * 1024 * 1024, VolumeClass: "hdd-archive", CreatedAt: demoNow2d},
		{Name: "backup-logs", Namespace: "default", State: "Pending", SizeBytes: 200 * 1024 * 1024 * 1024, VolumeClass: "hdd-archive", CreatedAt: demoNow1h},
	},
	"volumesnapshots": []VolumeSnapshotResponse{
		{Name: "data-01-snap-2026-10-31", Namespace: "default", VolumeRef: "data-01", State: "Ready", CreatedAt: demoNow1d},
		{Name: "api-data-snap-2026-10-30", Namespace: "default", VolumeRef: "api-data", State: "Ready", CreatedAt: demoNow2d},
	},

	"networks": []NetworkResponse{
		{Name: "prod-network", Namespace: "default", CreatedAt: demoNow2d},
		{Name: "dev-network", Namespace: "default", CreatedAt: demoNow1d},
	},
	"networkinterfaces": []NetworkInterfaceResponse{
		{Name: "web-frontend-01-nic0", Namespace: "default", State: "Available", IPs: []string{"10.0.0.10"}, Network: "prod-network", Machine: "web-frontend-01", CreatedAt: demoNow2d},
		{Name: "web-frontend-02-nic0", Namespace: "default", State: "Available", IPs: []string{"10.0.0.11"}, Network: "prod-network", Machine: "web-frontend-02", CreatedAt: demoNow2d},
		{Name: "api-server-01-nic0", Namespace: "default", State: "Available", IPs: []string{"10.0.0.20"}, Network: "prod-network", Machine: "api-server-01", CreatedAt: demoNow1d},
	},
	"natgateways": []NATGatewayResponse{
		{Name: "prod-nat-gw", Namespace: "default", NetworkRef: "prod-network", IPs: []string{"203.0.113.5"}, CreatedAt: demoNow2d},
	},

	"virtualips": []VirtualIPResponse{
		{Name: "ingress-vip-01", Namespace: "default", IP: "203.0.113.10", Type: "Public", IPFamily: "IPv4", CreatedAt: demoNow2d},
		{Name: "api-vip-01", Namespace: "default", IP: "203.0.113.11", Type: "Public", IPFamily: "IPv4", CreatedAt: demoNow1d},
	},
	"loadbalancers": []LoadBalancerResponse{
		{Name: "web-lb", Namespace: "default", Type: "Public", IPs: []string{"203.0.113.20"}, CreatedAt: demoNow2d},
		{Name: "api-lb", Namespace: "default", Type: "Internal", IPs: []string{"10.0.1.100"}, CreatedAt: demoNow1d},
	},

	"prefixes": []PrefixResponse{
		{Name: "prod-cidr", Namespace: "default", Prefix: "10.0.0.0/16", Phase: "Allocated", CreatedAt: demoNow2d},
		{Name: "dev-cidr", Namespace: "default", Prefix: "10.1.0.0/16", Phase: "Allocated", CreatedAt: demoNow1d},
		{Name: "staging-cidr", Namespace: "default", Prefix: "10.2.0.0/16", Phase: "Pending", CreatedAt: demoNow1h},
	},
	"prefixallocations": []PrefixAllocationResponse{
		{Name: "web-frontend-01-alloc", Namespace: "default", PrefixRef: "prod-cidr", Prefix: "10.0.0.0/28", Phase: "Allocated", CreatedAt: demoNow2d},
		{Name: "api-server-01-alloc", Namespace: "default", PrefixRef: "prod-cidr", Prefix: "10.0.0.16/28", Phase: "Allocated", CreatedAt: demoNow1d},
	},
	"resourcequotas": []ResourceQuotaResponse{
		{
			Name: "default-quota", Namespace: "default",
			Hard: map[string]string{"compute.ironcore.dev/machines": "20", "storage.ironcore.dev/requests.storage": "5Ti"},
			Used: map[string]string{"compute.ironcore.dev/machines": "5", "storage.ironcore.dev/requests.storage": "1Ti"},
			CreatedAt: demoNow2d,
		},
	},

	// Phase 4
	"networkpolicies": []NetworkPolicyResponse{
		{Name: "allow-web-ingress", Namespace: "default", NetworkRef: "prod-network", PolicyTypes: []string{"Ingress"}, IngressRules: 2, EgressRules: 0, CreatedAt: demoNow1d},
		{Name: "restrict-egress", Namespace: "default", NetworkRef: "prod-network", PolicyTypes: []string{"Egress"}, IngressRules: 0, EgressRules: 1, CreatedAt: demoNow2h},
	},
	"buckets": []BucketResponse{
		{Name: "assets-bucket", Namespace: "default", BucketClassRef: "standard-s3", BucketPoolRef: "bucket-pool-eu", State: "Available", Endpoint: "https://assets-bucket.storage.example.com", CreatedAt: demoNow2d},
		{Name: "backup-bucket", Namespace: "default", BucketClassRef: "standard-s3", BucketPoolRef: "bucket-pool-eu", State: "Available", Endpoint: "https://backup-bucket.storage.example.com", CreatedAt: demoNow1d},
		{Name: "ml-data-bucket", Namespace: "default", BucketClassRef: "fast-s3", BucketPoolRef: "bucket-pool-eu", State: "Pending", Endpoint: "", CreatedAt: demoNow1h},
	},
	"bucketclasses": []BucketClassResponse{
		{Name: "standard-s3", Capabilities: map[string]string{"tps": "1000", "iops": "3000"}, CreatedAt: demoNow2d},
		{Name: "fast-s3", Capabilities: map[string]string{"tps": "5000", "iops": "16000"}, CreatedAt: demoNow2d},
	},
	"bucketpools": []BucketPoolResponse{
		{Name: "bucket-pool-eu", State: "Available", AvailableBucketClasses: []string{"standard-s3", "fast-s3"}, CreatedAt: demoNow2d},
	},

	"events": []EventResponse{
		{Namespace: "default", Name: "web-frontend-01.17a1b", Reason: "Starting", Message: "Starting container runtime", Type: "Normal", Count: 1, FirstTime: demoNow2d, LastTime: demoNow2d, Regarding: "web-frontend-01", RegardsKind: "Machine"},
		{Namespace: "default", Name: "web-frontend-01.17a2c", Reason: "Started", Message: "Machine started successfully", Type: "Normal", Count: 1, FirstTime: demoNow2d, LastTime: demoNow2d, Regarding: "web-frontend-01", RegardsKind: "Machine"},
		{Namespace: "default", Name: "api-server-01.18a3d", Reason: "Provisioning", Message: "Waiting for volume attachment", Type: "Normal", Count: 3, FirstTime: demoNow1d, LastTime: demoNow1h, Regarding: "api-server-01", RegardsKind: "Machine"},
		{Namespace: "default", Name: "data-01.19a1e", Reason: "Bound", Message: "Volume bound to node", Type: "Normal", Count: 1, FirstTime: demoNow2d, LastTime: demoNow2d, Regarding: "data-01", RegardsKind: "Volume"},
		{Namespace: "default", Name: "worker-batch-01.1aa2f", Reason: "FailedScheduling", Message: "0/2 nodes are available: insufficient CPU", Type: "Warning", Count: 5, FirstTime: demoNow1h, LastTime: demoNow, Regarding: "worker-batch-01", RegardsKind: "Machine"},
	},
}

// MachineClassResponse is here since it's not in types.go yet.
type MachineClassResponse struct {
	Name string `json:"name"`
	CPU  string `json:"cpu"`
	RAM  string `json:"ram"`
}

func (d *DemoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// chi's r.Mount strips the mount prefix, so path starts after /api/v1.
	// We also handle the full path in case called directly.
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/v1")
	if path == "" {
		path = "/"
	}

	// namespaces
	if path == "/namespaces" {
		writeJSON(w, http.StatusOK, demoFixtures["namespaces"])
		return
	}
	// machineclasses
	if path == "/machineclasses" {
		writeJSON(w, http.StatusOK, demoFixtures["machineclasses"])
		return
	}
	// volumeclasses
	if path == "/volumeclasses" {
		writeJSON(w, http.StatusOK, demoFixtures["volumeclasses"])
		return
	}
	// phase 3 cluster-scoped
	if path == "/machinepools" {
		writeJSON(w, http.StatusOK, demoFixtures["machinepools"])
		return
	}
	if path == "/volumepools" {
		writeJSON(w, http.StatusOK, demoFixtures["volumepools"])
		return
	}
	// Phase 4 cluster-scoped
	if path == "/bucketclasses" {
		writeJSON(w, http.StatusOK, demoFixtures["bucketclasses"])
		return
	}
	if path == "/bucketpools" {
		writeJSON(w, http.StatusOK, demoFixtures["bucketpools"])
		return
	}

	// namespace-scoped routes: /namespaces/{ns}/resource[/{name}]
	if strings.HasPrefix(path, "/namespaces/") {
		parts := strings.Split(strings.TrimPrefix(path, "/namespaces/"), "/")
		if len(parts) < 2 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		// parts[0] = ns, parts[1] = resource, parts[2] = name (optional)
		resource := parts[1]
		name := ""
		if len(parts) > 2 {
			name = parts[2]
		}
		// handle power patch
		if len(parts) > 3 && parts[3] == "power" && r.Method == http.MethodPatch {
			var req struct{ Power string `json:"power"` }
			json.NewDecoder(r.Body).Decode(&req)
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}
		d.handleResource(w, r, resource, name)
		return
	}

	writeError(w, http.StatusNotFound, "demo: unknown route "+path)
}

func (d *DemoHandler) handleResource(w http.ResponseWriter, r *http.Request, resource, name string) {
	if r.Method == http.MethodPost {
		// Echo back a synthetic created response
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		body["createdAt"] = demoNow
		body["namespace"] = "default"
		writeJSON(w, http.StatusCreated, body)
		return
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch resource {
	case "machines":
		list := demoFixtures["machines"].([]MachineResponse)
		if name == "" {
			writeJSON(w, http.StatusOK, list)
		} else {
			for _, m := range list {
				if m.Name == name {
					writeJSON(w, http.StatusOK, VolumeDetailResponse{
						Name: m.Name, Namespace: m.Namespace,
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "machine not found: "+name)
		}
	case "volumes":
		list := demoFixtures["volumes"].([]VolumeResponse)
		if name == "" {
			writeJSON(w, http.StatusOK, list)
		} else {
			for _, v := range list {
				if v.Name == name {
					writeJSON(w, http.StatusOK, VolumeDetailResponse{
						Name: v.Name, Namespace: v.Namespace, VolumeClass: v.VolumeClass,
						SizeGiB: v.SizeBytes / (1024 * 1024 * 1024),
						State: v.State, VolumeID: "vol-demo-" + v.Name,
						AccessDriver: "csi", CreatedAt: v.CreatedAt,
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "volume not found: "+name)
		}
	case "networks":
		list := demoFixtures["networks"].([]NetworkResponse)
		if name == "" {
			writeJSON(w, http.StatusOK, list)
		} else {
			for _, n := range list {
				if n.Name == name {
					writeJSON(w, http.StatusOK, NetworkDetailResponse{
						Name: n.Name, Namespace: n.Namespace,
						State: "Available", ProviderID: "demo-provider-" + n.Name,
						Peerings: []NetworkPeeringStatus{}, CreatedAt: n.CreatedAt,
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "network not found: "+name)
		}
	case "networkinterfaces":
		writeJSON(w, http.StatusOK, demoFixtures["networkinterfaces"])
	case "natgateways":
		if name == "" {
			writeJSON(w, http.StatusOK, demoFixtures["natgateways"])
		} else {
			writeError(w, http.StatusNotFound, "natgateway not found: "+name)
		}
	case "volumesnapshots":
		if name == "" {
			writeJSON(w, http.StatusOK, demoFixtures["volumesnapshots"])
		} else {
			writeError(w, http.StatusNotFound, "volumesnapshot not found: "+name)
		}
	case "prefixallocations":
		writeJSON(w, http.StatusOK, demoFixtures["prefixallocations"])
	case "resourcequotas":
		writeJSON(w, http.StatusOK, demoFixtures["resourcequotas"])
	case "networkpolicies":
		writeJSON(w, http.StatusOK, demoFixtures["networkpolicies"])
	case "buckets":
		writeJSON(w, http.StatusOK, demoFixtures["buckets"])
	case "events":
		regarding := r.URL.Query().Get("regarding")
		kind := r.URL.Query().Get("kind")
		all := demoFixtures["events"].([]EventResponse)
		filtered := make([]EventResponse, 0, len(all))
		for _, ev := range all {
			if regarding != "" && ev.Regarding != regarding {
				continue
			}
			if kind != "" && ev.RegardsKind != kind {
				continue
			}
			filtered = append(filtered, ev)
		}
		writeJSON(w, http.StatusOK, filtered)
	case "virtualips":
		list := demoFixtures["virtualips"].([]VirtualIPResponse)
		if name == "" {
			writeJSON(w, http.StatusOK, list)
		} else {
			for _, v := range list {
				if v.Name == name {
					writeJSON(w, http.StatusOK, VirtualIPDetailResponse{
						Name: v.Name, Namespace: v.Namespace, Type: v.Type,
						IPFamily: v.IPFamily, IP: v.IP, TargetRef: "", CreatedAt: v.CreatedAt,
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "virtualip not found: "+name)
		}
	case "loadbalancers":
		list := demoFixtures["loadbalancers"].([]LoadBalancerResponse)
		if name == "" {
			writeJSON(w, http.StatusOK, list)
		} else {
			for _, lb := range list {
				if lb.Name == name {
					writeJSON(w, http.StatusOK, LoadBalancerDetailResponse{
						Name: lb.Name, Namespace: lb.Namespace, Type: lb.Type,
						IPFamilies: []string{"IPv4"}, NetworkRef: "prod-network",
						Ports: []LBPort{{Protocol: "TCP", Port: 443}},
						IPs: lb.IPs, CreatedAt: lb.CreatedAt,
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "loadbalancer not found: "+name)
		}
	case "prefixes":
		list := demoFixtures["prefixes"].([]PrefixResponse)
		if name == "" {
			writeJSON(w, http.StatusOK, list)
		} else {
			for _, p := range list {
				if p.Name == name {
					writeJSON(w, http.StatusOK, PrefixDetailResponse{
						Name: p.Name, Namespace: p.Namespace,
						IPFamily: "IPv4", Prefix: p.Prefix, Phase: p.Phase,
						ParentRef: "", Used: []string{}, CreatedAt: p.CreatedAt,
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "prefix not found: "+name)
		}
	default:
		writeError(w, http.StatusNotFound, "demo: unknown resource: "+resource)
	}
}
