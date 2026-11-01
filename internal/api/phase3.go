package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachinePoolHandler serves MachinePool list (cluster-scoped)
type MachinePoolHandler struct{ cs versioned.Interface }

func NewMachinePoolHandler(cs versioned.Interface) *MachinePoolHandler {
	return &MachinePoolHandler{cs: cs}
}

func (h *MachinePoolHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.ComputeV1alpha1().MachinePools().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]MachinePoolResponse, 0, len(list.Items))
	for _, mp := range list.Items {
		classes := make([]string, 0, len(mp.Status.AvailableMachineClasses))
		for _, c := range mp.Status.AvailableMachineClasses {
			classes = append(classes, c.Name)
		}
		resp = append(resp, MachinePoolResponse{
			Name:                    mp.Name,
			State:                   string(mp.Status.State),
			AvailableMachineClasses: classes,
			CreatedAt:               mp.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// VolumePoolHandler serves VolumePool list (cluster-scoped)
type VolumePoolHandler struct{ cs versioned.Interface }

func NewVolumePoolHandler(cs versioned.Interface) *VolumePoolHandler {
	return &VolumePoolHandler{cs: cs}
}

func (h *VolumePoolHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.StorageV1alpha1().VolumePools().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]VolumePoolResponse, 0, len(list.Items))
	for _, vp := range list.Items {
		classes := make([]string, 0, len(vp.Status.AvailableVolumeClasses))
		for _, c := range vp.Status.AvailableVolumeClasses {
			classes = append(classes, c.Name)
		}
		resp = append(resp, VolumePoolResponse{
			Name:                   vp.Name,
			State:                  string(vp.Status.State),
			AvailableVolumeClasses: classes,
			CreatedAt:              vp.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// VolumeSnapshotHandler serves VolumeSnapshot CRUD in a namespace
type VolumeSnapshotHandler struct{ cs versioned.Interface }

func NewVolumeSnapshotHandler(cs versioned.Interface) *VolumeSnapshotHandler {
	return &VolumeSnapshotHandler{cs: cs}
}

func (h *VolumeSnapshotHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.StorageV1alpha1().VolumeSnapshots(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]VolumeSnapshotResponse, 0, len(list.Items))
	for _, vs := range list.Items {
		volRef := ""
		if vs.Spec.VolumeRef != nil {
			volRef = vs.Spec.VolumeRef.Name
		}
		resp = append(resp, VolumeSnapshotResponse{
			Name:      vs.Name,
			Namespace: vs.Namespace,
			VolumeRef: volRef,
			State:     string(vs.Status.State),
			CreatedAt: vs.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VolumeSnapshotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	err := h.cs.StorageV1alpha1().VolumeSnapshots(ns).Delete(r.Context(), name, metav1.DeleteOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// NATGatewayHandler serves NATGateway list/get/delete
type NATGatewayHandler struct{ cs versioned.Interface }

func NewNATGatewayHandler(cs versioned.Interface) *NATGatewayHandler {
	return &NATGatewayHandler{cs: cs}
}

func (h *NATGatewayHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().NATGateways(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]NATGatewayResponse, 0, len(list.Items))
	for _, ng := range list.Items {
		ips := make([]string, 0, len(ng.Status.IPs))
		for _, ip := range ng.Status.IPs {
			ips = append(ips, ip.String())
		}
		resp = append(resp, NATGatewayResponse{
			Name:       ng.Name,
			Namespace:  ng.Namespace,
			NetworkRef: ng.Spec.NetworkRef.Name,
			IPs:        ips,
			CreatedAt:  ng.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NATGatewayHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	err := h.cs.NetworkingV1alpha1().NATGateways(ns).Delete(r.Context(), name, metav1.DeleteOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PrefixAllocationHandler serves PrefixAllocation list per namespace
type PrefixAllocationHandler struct{ cs versioned.Interface }

func NewPrefixAllocationHandler(cs versioned.Interface) *PrefixAllocationHandler {
	return &PrefixAllocationHandler{cs: cs}
}

func (h *PrefixAllocationHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.IpamV1alpha1().PrefixAllocations(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]PrefixAllocationResponse, 0, len(list.Items))
	for _, pa := range list.Items {
		prefixStr := ""
		if pa.Spec.Prefix != nil {
			prefixStr = pa.Spec.Prefix.String()
		}
		prefixRef := ""
		if pa.Spec.PrefixRef != nil {
			prefixRef = pa.Spec.PrefixRef.Name
		}
		resp = append(resp, PrefixAllocationResponse{
			Name:      pa.Name,
			Namespace: pa.Namespace,
			PrefixRef: prefixRef,
			Prefix:    prefixStr,
			Phase:     string(pa.Status.Phase),
			CreatedAt: pa.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// ResourceQuotaHandler serves ResourceQuota list per namespace
type ResourceQuotaHandler struct{ cs versioned.Interface }

func NewResourceQuotaHandler(cs versioned.Interface) *ResourceQuotaHandler {
	return &ResourceQuotaHandler{cs: cs}
}

func (h *ResourceQuotaHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.CoreV1alpha1().ResourceQuotas(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]ResourceQuotaResponse, 0, len(list.Items))
	for _, rq := range list.Items {
		hard := make(map[string]string, len(rq.Spec.Hard))
		for k, v := range rq.Spec.Hard {
			hard[string(k)] = v.String()
		}
		used := make(map[string]string, len(rq.Status.Used))
		for k, v := range rq.Status.Used {
			used[string(k)] = v.String()
		}
		resp = append(resp, ResourceQuotaResponse{
			Name:      rq.Name,
			Namespace: rq.Namespace,
			Hard:      hard,
			Used:      used,
			CreatedAt: rq.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// Prevent unused import warnings for types imported for documentation
var _ = computev1alpha1.MachinePool{}
var _ = storagev1alpha1.VolumePool{}
var _ = networkingv1alpha1.NATGateway{}
var _ = ipamv1alpha1.PrefixAllocation{}
