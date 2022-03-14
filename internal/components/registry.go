package components

import (
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/grafana/grafana/pkg/schema"
)

// Bridge
type Bridge interface {
	ClientForSchema(schema schema.ObjectSchema) (rest.Interface, error)
	ControllerManager() ctrl.Manager
}

// Coremodel
type Coremodel interface {
	Schema() schema.ObjectSchema
	RegisterController(bridge Bridge) error
}

// Registry
type Registry struct {
	models []Coremodel
}

// NewRegistry
func NewRegistry(models ...Coremodel) *Registry {
	return &Registry{
		models: models,
	}
}

// Coremodels
func (r *Registry) Coremodels() []Coremodel {
	return r.models
}
