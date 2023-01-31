package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/henderiw-k8s-lcnc/fn-lcnc-fabric-image/pkg/ipam"
	"github.com/henderiw-k8s-lcnc/fn-sdk/go/fn"
	topov1alpha1 "github.com/henderiw-k8s-lcnc/topology/apis/topo/v1alpha1"
	"github.com/henderiw/fabric/fabric"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Fabric implements fn.Runner
var _ fn.Runner = &Fabric{}

type Fabric struct {
	definition      *topov1alpha1.Definition
	masterTemplates []*topov1alpha1.Template
	childTemplates  []*topov1alpha1.Template
}

func main() {
	ctx := context.TODO()
	if err := fn.AsMain(fn.WithContext(ctx, &Fabric{
		definition:      &topov1alpha1.Definition{}, // we expect only 1 defintion
		masterTemplates: []*topov1alpha1.Template{},
		childTemplates:  []*topov1alpha1.Template{},
	})); err != nil {
		os.Exit(1)
	}
}

// Run is the main function logic.
// `functionConfig` is from the STDIN "ResourceContext.FunctionConfig".
// `resources` is parsed from the STDIN "ResourceList.Resources".
// `results` provides easy methods to add info, error result to `ResourceContext.Results`.
func (r *Fabric) Run(ctx *fn.Context, functionConfig map[string]runtime.RawExtension, resources *fn.Resources, results *fn.Results) bool {
	// parse input
	for gvkString, gvkResources := range resources.Resources {
		for _, gvkResource := range gvkResources {
			switch gvkString {
			case "Definition.v1alpha1.topo.yndd.io":
				if err := json.Unmarshal(gvkResource.Raw, r.definition); err != nil {
					results.ErrorE(err)
				}
			case "Template.v1alpha1.topo.yndd.io":
				t := &topov1alpha1.Template{}
				if err := json.Unmarshal(gvkResource.Raw, t); err != nil {
					results.ErrorE(err)
				}
				if t.Spec.Properties.Fabric.HasReference() {
					r.masterTemplates = append(r.masterTemplates, t)
				} else {
					r.childTemplates = append(r.childTemplates, t)
				}
			}
		}
	}

	// create the fabric from the input
	f, err := fabric.New(&fabric.Config{
		Name:            r.definition.GetName(),
		Namespace:       r.definition.GetNamespace(),
		MasterTemplates: r.masterTemplates,
		ChildTemplates:  r.childTemplates,
		Location:        r.definition.Spec.Properties.Location,
	})
	if err != nil {
		results.ErrorE(err)
	}

	for _, n := range f.GetNodes() {
		resources.AddResource(n, &fn.ResourceParameters{})
	}
	for _, l := range f.GetLinks() {
		resources.AddResource(l, &fn.ResourceParameters{})
	}

	//allocate ip prefixes and ip(s) per node for mgmt purposes
	prefixAlloc := &ipam.IpamAllocInfo{
		Name:      r.definition.GetName(),
		Namespace: r.definition.GetNamespace(),
		Spec: ipamv1alpha1.IPAllocationSpec{
			PrefixKind: ipamv1alpha1.PrefixKindLoopback,
			NetworkInstanceRef: &ipamv1alpha1.NetworkInstanceReference{
				Name:      "vpc-mgmt-fabric", //should come from the definition
				Namespace: r.definition.GetNamespace(),
			},
			PrefixLength:  24,
			AddressFamily: "ipv4",
			CreatePrefix:  true,
			Labels: map[string]string{
				//ipamv1alpha1.NephioRegionKey:  "us-central-1",
				"nephio.org/region": "us-central-1",
				ipamv1alpha1.NephioSiteKey:    "edge1",
				ipamv1alpha1.NephioPurposeKey: "mgmt",
			},
		},
	}
	resources.AddResource(prefixAlloc.BuildIPAMIPPrefixAllocation(), &fn.ResourceParameters{Conditioned: true, Internal: true})
	for idx, n := range f.GetNodes() {
		ipAlloc := &ipam.IpamAllocInfo{
			Name:      n.GetName(),
			Namespace: r.definition.GetNamespace(),
			Spec: ipamv1alpha1.IPAllocationSpec{
				PrefixKind: ipamv1alpha1.PrefixKindLoopback,
				NetworkInstanceRef: &ipamv1alpha1.NetworkInstanceReference{
					Name:      "vpc-mgmt-fabric", //should come from the definition
					Namespace: r.definition.GetNamespace(),
				},
				AddressFamily: "ipv4",
				Index:         uint32(idx),
				Labels: map[string]string{
					//ipamv1alpha1.NephioRegionKey:  "us-central-1",
					"nephio.org/region": "us-central-1",
					ipamv1alpha1.NephioSiteKey:    "edge1",
					ipamv1alpha1.NephioPurposeKey: "mgmt",
				},
				//DependsOn:           prefixAlloc.GetName(),
			},
		}
		resources.AddResource(ipAlloc.BuildIPAMIPAllocation(), &fn.ResourceParameters{Conditioned: true, Internal: true})
	}

	return true
}
