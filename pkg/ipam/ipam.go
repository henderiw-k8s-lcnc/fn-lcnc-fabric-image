package ipam

import (
	"fmt"

	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IpamAllocInfo struct {
	Name      string
	Namespace string
	Spec      ipamv1alpha1.IPAllocationSpec
}

func (r *IpamAllocInfo) BuildIPAMIPAllocation() *ipamv1alpha1.IPAllocation {
	return &ipamv1alpha1.IPAllocation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ipamv1alpha1.GroupVersion.String(),
			Kind:       ipamv1alpha1.IPAllocationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.GetName(),
			Namespace: r.Namespace,
			Labels:    r.GetLabels(),
		},
		Spec: r.Spec,
	}
}

func (r *IpamAllocInfo) BuildIPAMIPPrefixAllocation() *ipamv1alpha1.IPAllocation {
	return &ipamv1alpha1.IPAllocation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ipamv1alpha1.GroupVersion.String(),
			Kind:       ipamv1alpha1.IPAllocationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.GetName(),
			Namespace: r.Namespace,
			Labels:    r.GetLabels(),
		},
		Spec: r.Spec,
	}
}

func (r *IpamAllocInfo) GetName() string {
	//if az, ok := r.Spec.Labels[ipamv1alpha1.NephioAvailabilityZoneKey]; ok {
	if az, ok := r.Spec.Labels["nephio.org/availability-zone"]; ok {
		return fmt.Sprintf("%s.%s.%s.%s.%s",
			r.Name,
			//r.Spec.Labels[ipamv1alpha1.NephioRegionKey],
			r.Spec.Labels["nephio.org/region"],
			az,
			r.Spec.NetworkInstanceRef.Name,
			r.Spec.Labels[ipamv1alpha1.NephioPurposeKey],
		)
	}
	return fmt.Sprintf("%s.%s.%s.%s",
		r.Name,
		//r.Spec.Labels[ipamv1alpha1.NephioRegionKey],
		r.Spec.Labels["nephio.org/region"],
		r.Spec.NetworkInstanceRef.Name,
		r.Spec.Labels[ipamv1alpha1.NephioPurposeKey],
	)
}

func (r *IpamAllocInfo) GetLabels() map[string]string {
	labels := map[string]string{}
	/*
		if r.DependsOn != "" {
			labels["dependsOn"] = r.DependsOn
		}
	*/
	return labels
}
