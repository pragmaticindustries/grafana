package schema

import (
	"io/fs"

	"github.com/grafana/thema"
	"github.com/grafana/thema/kernel"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ThemaSchema contains a Grafana schema where the canonical schema expression
// is made with Thema and CUE.
// TODO: figure out what fields should be here
type ThemaSchema struct {
	lineage        thema.Lineage
	groupName      string
	groupVersion   string
	openapiSchema  apiextensionsv1.JSONSchemaProps
	runtimeObjects []runtime.Object
}

// NewThemaSchema
// TODO: support multiple versions. Should be possible, since versions are in the lineage.
func NewThemaSchema(
	lineage thema.Lineage,
	groupName, groupVersion string, // TODO: somehow figure this out from the lineage
	openapiSchema apiextensionsv1.JSONSchemaProps, // TODO: should be part of the lineage
	resource, list runtime.Object,
) *ThemaSchema {
	return &ThemaSchema{
		lineage:        lineage,
		groupName:      groupName,
		groupVersion:   groupVersion,
		openapiSchema:  openapiSchema,
		runtimeObjects: []runtime.Object{resource, list},
	}
}

func LoadThemaSchema(
	path string,
	fs fs.FS,
	lib thema.Library,
	version thema.SyntacticVersion,
	object interface{},
	groupName, groupVersion string, // TODO: somehow figure this out from the lineage
	openapiSchema apiextensionsv1.JSONSchemaProps, // TODO: should be part of the lineage
	resource, list runtime.Object,
) (*ThemaSchema, error) {
	lin, err := LoadLineage(path, fs, lib)
	if err != nil {
		return nil, err
	}

	// Calling this ensures our program cannot start,
	// if the Go DataSource.Model type is not aligned with the canonical schema version in the lineage.
	if _, err := newJSONKernel(lin, path, object); err != nil {
		return nil, err
	}

	zsch, err := lin.Schema(version)
	if err != nil {
		return nil, err
	}

	if err := thema.AssignableTo(zsch, object); err != nil {
		return nil, err
	}

	return NewThemaSchema(
		lin,
		groupName,
		groupVersion,
		openapiSchema,
		resource,
		list,
	), nil
}

// Name returns the canonical string that identifies the object being schematized.
func (ts ThemaSchema) Name() string {
	return ts.lineage.Name()
}

// GroupName
func (ts ThemaSchema) GroupName() string {
	return ts.groupName
}

// GroupName
func (ts ThemaSchema) GroupVersion() string {
	return ts.groupVersion
}

// GetRuntimeObjects returns a runtime.Object that will accurately represent
// the authorial intent of the Thema lineage to Kubernetes.
func (ts ThemaSchema) RuntimeObjects() []runtime.Object {
	return ts.runtimeObjects
}

// OpenAPISchema
func (ts ThemaSchema) OpenAPISchema() apiextensionsv1.JSONSchemaProps {
	return ts.openapiSchema
}

func newJSONKernel(lin thema.Lineage, loaderPath string, object interface{}) (kernel.InputKernel, error) {
	return kernel.NewInputKernel(kernel.InputKernelConfig{
		Lineage:     lin,
		Loader:      kernel.NewJSONDecoder(loaderPath),
		To:          thema.SV(0, 0),
		TypeFactory: func() interface{} { return object },
	})
}
