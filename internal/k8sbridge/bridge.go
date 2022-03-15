// Package k8sbridge provides interfaces for communicating with an underlying
// Kubernetes apiserver

package k8sbridge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	k8schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	"github.com/grafana/grafana/internal/components"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/schema"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	groupName = "grafana.core.group"
	// TODO come up with rule governing when and why this is incremented
	groupVersion = "v1alpha1"
)

// Service
type Service struct {
	config  *rest.Config
	client  *Clientset
	schemas schema.CoreSchemaList
	manager ctrl.Manager
	enabled bool
	logger  log.Logger
}

// CoremodelRegistry
type CoremodelRegistry interface {
	Coremodels() []components.Coremodel
}

// ProvideService
func ProvideService(cfg *setting.Cfg, features featuremgmt.FeatureToggles, registry CoremodelRegistry) (*Service, error) {
	enabled := features.IsEnabled(featuremgmt.FlagIntentapi)
	if !enabled {
		return &Service{
			enabled: false,
		}, nil
	}

	sec := cfg.Raw.Section("intentapi.kubebridge")
	configPath := sec.Key("kubeconfig_path").MustString("")

	if configPath == "" {
		return nil, errors.New("kubeconfig path cannot be empty when using Intent API")
	}

	configPath = filepath.Clean(configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cannot find kubeconfig file at '%s'", configPath)
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	schm := runtime.NewScheme()

	models := registry.Coremodels()
	schemas := make(schema.CoreSchemaList, 0, len(models))
	for _, m := range models {
		s := m.Schema()

		schemaBuilder := &scheme.Builder{
			GroupVersion: k8schema.GroupVersion{
				Group:   s.GroupName(),
				Version: s.GroupVersion(),
			},
		}

		schemaBuilder.Register(s.RuntimeObjects()...)

		if err := schemaBuilder.AddToScheme(schm); err != nil {
			return nil, err
		}

		schemas = append(schemas, s)
	}

	// TODO: pass models to clientset to create clients and register CRDs.
	cset, err := NewClientset(config, schemas)
	if err != nil {
		return nil, err
	}

	fmt.Println("scheme", schm)

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: schm,
	})
	if err != nil {
		return nil, err
	}

	for _, m := range models {
		if err := m.RegisterController(mgr); err != nil {
			return nil, err
		}
	}

	return &Service{
		config:  config,
		client:  cset,
		schemas: schemas,
		manager: mgr,
		enabled: enabled,
		logger:  log.New("k8sbridge.service"),
	}, nil
}

// IsDisabled
func (s *Service) IsDisabled() bool {
	return !s.enabled
}

// Run
func (s *Service) Run(ctx context.Context) error {
	fmt.Println("starting k8s service")
	if err := s.manager.Start(ctx); err != nil {
		return err
	}

	return nil
}

// RestConfig
func (s *Service) RestConfig() *rest.Config {
	return s.config
}

// ControllerManager
func (s *Service) ControllerManager() ctrl.Manager {
	return s.manager
}
