package datasource

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/grafana/grafana/internal/components"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/schema"
	"github.com/grafana/thema"
)

// Store
//
// TODO: I think we should define a generic store interface similar to k8s rest.Interface
// and have storeset around (similar to clientset) from which we can grab specific store implementation for schema.
type Store interface {
	Get(ctx context.Context, uid string) (ModelObject, error)
	Insert(ctx context.Context, ds ModelObject) error
	Update(ctx context.Context, ds ModelObject) error
	Delete(ctx context.Context, uid string) error
}

// Coremodel
type Coremodel struct {
	store  Store
	client rest.Interface
	schema schema.ObjectSchema
}

// ProvideCoremodel
func ProvideCoremodel(store Store, schemaLib thema.Library) (*Coremodel, error) {
	schema, err := schema.LoadThemaSchema(
		cuePath,
		cueFS,
		schemaLib,
		schemaVersion,
		&ModelSpec{},
		groupName,
		groupVersion,
		schemaOpenapi,
		&ModelObject{},
		&ModelObjectList{},
	)
	if err != nil {
		return nil, err
	}

	return &Coremodel{
		store:  store,
		schema: schema,
	}, nil
}

// Schema
func (m *Coremodel) Schema() schema.ObjectSchema {
	return m.schema
}

// RegisterController
func (m *Coremodel) RegisterController(bridge components.Bridge) error {
	cli, err := bridge.ClientForSchema(m.schema)
	if err != nil {
		return err
	}

	if err := ctrl.
		NewControllerManagedBy(bridge.ControllerManager()).
		Named("datasources-controller").
		For(&ModelObject{}).
		Complete(m); err != nil {
		return err
	}

	m.client = cli
	return nil
}

// Reconcile
func (m *Coremodel) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	fmt.Println("Got for reconciliation", req.String())

	ds := ModelObject{}
	err := m.client.Get().Namespace(req.Namespace).Resource("datasources").Name(req.Name).Do(ctx).Into(&ds)

	// TODO: check ACTUAL error
	if errors.Is(err, rest.ErrNotInCluster) {
		return reconcile.Result{}, m.store.Delete(ctx, req.Name)
	}

	if err != nil {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 1 * time.Minute,
		}, err
	}

	_, err = m.store.Get(ctx, string(ds.UID))
	if err != nil {
		if !errors.Is(err, models.ErrDataSourceNotFound) {
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: 1 * time.Minute,
			}, err
		}

		if err := m.store.Insert(ctx, ds); err != nil {
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: 1 * time.Minute,
			}, err
		}
	}

	if err := m.store.Update(ctx, ds); err != nil {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 1 * time.Minute,
		}, err
	}

	return reconcile.Result{}, nil
}
