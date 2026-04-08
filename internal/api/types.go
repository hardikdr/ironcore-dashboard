package api

import (
	"encoding/json"
	"net/http"
)

type MachineResponse struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	State        string   `json:"state"`
	Power        string   `json:"power"`
	MachineClass string   `json:"machineClass"`
	Image        string   `json:"image"`
	IPs          []string `json:"ips"`
	Volumes      []string `json:"volumes"`
	CreatedAt    string   `json:"createdAt"`
}

type VolumeResponse struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	State       string `json:"state"`
	SizeBytes   int64  `json:"sizeBytes"`
	VolumeClass string `json:"volumeClass"`
	CreatedAt   string `json:"createdAt"`
}

type NetworkResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	CreatedAt string `json:"createdAt"`
}

type NetworkInterfaceResponse struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	State     string   `json:"state"`
	IPs       []string `json:"ips"`
	Network   string   `json:"network"`
	Machine   string   `json:"machine"`
	CreatedAt string   `json:"createdAt"`
}

type VirtualIPResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	IP        string `json:"ip"`
	Type      string `json:"type"`
	IPFamily  string `json:"ipFamily"`
	CreatedAt string `json:"createdAt"`
}

type LoadBalancerResponse struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Type      string   `json:"type"`
	IPs       []string `json:"ips"`
	CreatedAt string   `json:"createdAt"`
}

type PrefixResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Prefix    string `json:"prefix"`
	Phase     string `json:"phase"`
	CreatedAt string `json:"createdAt"`
}

type CreateMachineRequest struct {
	Name         string             `json:"name"`
	MachineClass string             `json:"machineClass"`
	Image        string             `json:"image"`
	NetworkName  string             `json:"networkName"`
	Volumes      []VolumeAttachment `json:"volumes"`
	Power        string             `json:"power"`
}

type VolumeAttachment struct {
	Name        string `json:"name"`
	SizeBytes   int64  `json:"sizeBytes"`
	VolumeClass string `json:"volumeClass"`
}

type PatchPowerRequest struct {
	Power string `json:"power"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
