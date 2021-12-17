package output

import (
	"encoding/json"
	"os"
	"sort"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/resource"
)

const TerraformOutputType = "terraform"

type Terraform struct {
	path string
}

func NewTerraform(path string) *Terraform {
	return &Terraform{path}
}

func (t *Terraform) newResource(res *resource.Resource) *tfjson.StateResource {

	source := res.Source
	if source == nil {
		source = &resource.TerraformStateSource{
			Ty:   res.ResourceType(),
			Name: res.ResourceId(),
		}
	}
	var schemaVersion uint64
	if res.Sch != nil {
		schemaVersion = uint64(res.Sch.SchemaVersion)
	}
	var attrs map[string]interface{}
	if res.Attrs != nil {
		attrs = *res.Attrs
	}

	return &tfjson.StateResource{
		Mode:            tfjson.ManagedResourceMode,
		Index:           source.Index(),
		ProviderName:    "not_implemented",
		Address:         source.Address(),
		Type:            res.ResourceType(),
		Name:            source.InternalName(),
		SchemaVersion:   schemaVersion,
		AttributeValues: attrs,
	}
}

func (t Terraform) Write(analysis *analyser.Analysis) error {
	var rootModuleResources []*tfjson.StateResource
	var modules []*tfjson.StateModule

	for _, res := range analysis.Managed() {
		if res.Src().Namespace() != "" {
			var module *tfjson.StateModule
			for _, m := range modules {
				if m.Address == res.Src().Namespace() {
					module = m
					break
				}
			}
			if module == nil {
				module = &tfjson.StateModule{Address: res.Src().Namespace()}
				modules = append(modules, module)
			}

			module.Resources = append(module.Resources, t.newResource(res))
			continue
		}
		rootModuleResources = append(rootModuleResources, t.newResource(res))
	}

	for _, res := range analysis.Unmanaged() {
		rootModuleResources = append(rootModuleResources, t.newResource(res))
	}

	for _, res := range analysis.Deleted() {
		rootModuleResources = append(rootModuleResources, t.newResource(res))
	}

	sort.SliceStable(rootModuleResources, func(i, j int) bool {
		return rootModuleResources[i].Address < rootModuleResources[j].Address
	})

	state := tfjson.State{
		FormatVersion:    "1.0",
		TerraformVersion: "not_implemented",
		Values: &tfjson.StateValues{
			RootModule: &tfjson.StateModule{
				Resources:    rootModuleResources,
				ChildModules: modules,
			},
		},
	}

	file := os.Stdout
	if !isStdOut(t.path) {
		f, err := os.OpenFile(t.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		file = f
	}
	result, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if _, err := file.Write(result); err != nil {
		return err
	}
	return nil

}
