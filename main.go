package main

import (
	"context"
	"encoding/json"
	"os"

	topov1alpha1 "github.com/henderiw-k8s-lcnc/topology/apis/topo/v1alpha1"
	"github.com/henderiw/fabric/fabric"
	"github.com/yndd/lcnc-function-sdk/go/fn"
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
	for gvkString, gvkResources := range resources.Input {
		for _, gvkResource := range gvkResources {
			switch gvkString {
			case "topo.yndd.io.v1alpha1.Definition":
				if err := json.Unmarshal(gvkResource.Raw, r.definition); err != nil {
					results.ErrorE(err)
				}
			case "topo.yndd.io.v1alpha1.Template":
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
		resources.AddOutput(n)
	}
	for _, l := range f.GetLinks() {
		resources.AddOutput(l)
	}

	return true
}
