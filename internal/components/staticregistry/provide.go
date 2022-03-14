package staticregistry

import (
	"github.com/grafana/grafana/internal/components"
	"github.com/grafana/grafana/internal/components/datasource"
)

// ProvideRegistry
func ProvideRegistry(
	datasourceModel *datasource.Coremodel,
) *components.Registry {
	return components.NewRegistry(
		datasourceModel,
	)
}
