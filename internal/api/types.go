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
	IP           string             `json:"ip"`
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

type VolumeClassResponse struct {
	Name    string `json:"name"`
	Storage string `json:"storage"`
}

type CreateVolumeRequest struct {
	Name             string `json:"name"`
	VolumeClass      string `json:"volumeClass"`
	SizeGiB          int64  `json:"sizeGiB"`
	EncryptionSecret string `json:"encryptionSecret,omitempty"`
}

type VolumeDetailResponse struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	VolumeClass  string `json:"volumeClass"`
	SizeGiB      int64  `json:"sizeGiB"`
	State        string `json:"state"`
	VolumeID     string `json:"volumeID"`
	AccessDriver string `json:"accessDriver"`
	CreatedAt    string `json:"createdAt"`
}

type CreateNetworkRequest struct {
	Name string `json:"name"`
}

type NetworkPeeringStatus struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type NetworkDetailResponse struct {
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	State      string                 `json:"state"`
	ProviderID string                 `json:"providerID"`
	Peerings   []NetworkPeeringStatus `json:"peerings"`
	CreatedAt  string                 `json:"createdAt"`
}

type CreateVirtualIPRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IPFamily string `json:"ipFamily"`
}

type VirtualIPDetailResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	IPFamily  string `json:"ipFamily"`
	IP        string `json:"ip"`
	TargetRef string `json:"targetRef"`
	CreatedAt string `json:"createdAt"`
}

type LBPort struct {
	Protocol string `json:"protocol"`
	Port     int32  `json:"port"`
	EndPort  int32  `json:"endPort,omitempty"`
}

type CreateLoadBalancerRequest struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	IPFamilies []string `json:"ipFamilies"`
	NetworkRef string   `json:"networkRef"`
	Ports      []LBPort `json:"ports"`
}

type LoadBalancerDetailResponse struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Type       string   `json:"type"`
	IPFamilies []string `json:"ipFamilies"`
	NetworkRef string   `json:"networkRef"`
	Ports      []LBPort `json:"ports"`
	IPs        []string `json:"ips"`
	CreatedAt  string   `json:"createdAt"`
}

type CreatePrefixRequest struct {
	Name         string `json:"name"`
	IPFamily     string `json:"ipFamily"`
	Prefix       string `json:"prefix,omitempty"`
	PrefixLength int32  `json:"prefixLength,omitempty"`
	ParentRef    string `json:"parentRef,omitempty"`
}

type PrefixDetailResponse struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	IPFamily  string   `json:"ipFamily"`
	Prefix    string   `json:"prefix"`
	Phase     string   `json:"phase"`
	ParentRef string   `json:"parentRef"`
	Used      []string `json:"used"`
	CreatedAt string   `json:"createdAt"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// WriteJSON and WriteError are exported for use outside this package (e.g. server.go).
func WriteJSON(w http.ResponseWriter, status int, v any) { writeJSON(w, status, v) }
func WriteError(w http.ResponseWriter, status int, msg string) { writeError(w, status, msg) }
