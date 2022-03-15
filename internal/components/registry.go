package components

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/grafana/grafana/pkg/schema"
)

// Coremodel
type Coremodel interface {
	Schema() schema.ObjectSchema
	RegisterController(ctrl.Manager) error
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
