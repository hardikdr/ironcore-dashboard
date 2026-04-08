package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	computev1alpha1    "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1       "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1    "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	versioned          "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1             "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1             "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MachineHandler struct{ cs versioned.Interface }

func NewMachineHandler(cs versioned.Interface) *MachineHandler {
	return &MachineHandler{cs: cs}
}

func (h *MachineHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.ComputeV1alpha1().Machines(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]MachineResponse, 0, len(list.Items))
	for _, m := range list.Items {
		resp = append(resp, machineToResponse(h.cs, &m, r))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *MachineHandler) Get(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	m, err := h.cs.ComputeV1alpha1().Machines(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, machineToResponse(h.cs, m, r))
}

func (h *MachineHandler) Create(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req CreateMachineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	power := computev1alpha1.Power("On")
	if req.Power == "Off" {
		power = computev1alpha1.Power("Off")
	}

	volumes := make([]computev1alpha1.Volume, 0, len(req.Volumes))
	for _, v := range req.Volumes {
		volumes = append(volumes, computev1alpha1.Volume{
			Name: v.Name,
			VolumeSource: computev1alpha1.VolumeSource{
				Ephemeral: &computev1alpha1.EphemeralVolumeSource{
					VolumeTemplate: &storagev1alpha1.VolumeTemplateSpec{
						Spec: storagev1alpha1.VolumeSpec{
							VolumeClassRef: &corev1.LocalObjectReference{Name: v.VolumeClass},
							Resources: corev1alpha1.ResourceList{
								corev1alpha1.ResourceStorage: *resource.NewQuantity(v.SizeBytes, resource.BinarySI),
							},
						},
					},
				},
			},
		})
	}

	machine := &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef: corev1.LocalObjectReference{Name: req.MachineClass},
			Power:           power,
			Volumes:         volumes,
			NetworkInterfaces: []computev1alpha1.NetworkInterface{
				{
					Name: "primary",
					NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
						Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
							NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
								Spec: networkingv1alpha1.NetworkInterfaceSpec{
									NetworkRef: corev1.LocalObjectReference{Name: req.NetworkName},
									IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
								},
							},
						},
					},
				},
			},
		},
	}

	created, err := h.cs.ComputeV1alpha1().Machines(ns).Create(r.Context(), machine, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, machineToResponse(h.cs, created, r))
}

func (h *MachineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.ComputeV1alpha1().Machines(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MachineHandler) PatchPower(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	var req PatchPowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	m, err := h.cs.ComputeV1alpha1().Machines(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	m.Spec.Power = computev1alpha1.Power(req.Power)
	updated, err := h.cs.ComputeV1alpha1().Machines(ns).Update(r.Context(), m, metav1.UpdateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, machineToResponse(h.cs, updated, r))
}

func (h *MachineHandler) ListMachineClasses(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.ComputeV1alpha1().MachineClasses().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	type MCResponse struct {
		Name string `json:"name"`
		CPU  string `json:"cpu"`
		RAM  string `json:"ram"`
	}
	resp := make([]MCResponse, 0, len(list.Items))
	for _, mc := range list.Items {
		cpu := mc.Capabilities.CPU().String()
		ram := mc.Capabilities.Memory().String()
		resp = append(resp, MCResponse{Name: mc.Name, CPU: cpu, RAM: ram})
	}
	writeJSON(w, http.StatusOK, resp)
}

// machineToResponse converts a Machine to a MachineResponse, fetching NIC IPs when available.
func machineToResponse(cs versioned.Interface, m *computev1alpha1.Machine, r *http.Request) MachineResponse {
	ips := []string{}
	for _, niStatus := range m.Status.NetworkInterfaces {
		if niStatus.NetworkInterfaceRef.Name != "" {
			nic, err := cs.NetworkingV1alpha1().NetworkInterfaces(m.Namespace).Get(r.Context(), niStatus.NetworkInterfaceRef.Name, metav1.GetOptions{})
			if err == nil {
				for _, ip := range nic.Status.IPs {
					ips = append(ips, ip.String())
				}
			}
		}
	}
	vols := []string{}
	for _, v := range m.Spec.Volumes {
		vols = append(vols, v.Name)
	}
	return MachineResponse{
		Name:         m.Name,
		Namespace:    m.Namespace,
		State:        string(m.Status.State),
		Power:        string(m.Spec.Power),
		MachineClass: m.Spec.MachineClassRef.Name,
		IPs:          ips,
		Volumes:      vols,
		CreatedAt:    m.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
