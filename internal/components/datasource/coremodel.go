package datasource

import (
	"context"
	"errors"
	"fmt"
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/schema"
	"github.com/grafana/thema"
)

// Store
//
// TODO: I think we should define a generic store interface similar to k8s rest.Interface
// and have storeset around (similar to clientset) from which we can grab specific store implementation for schema.
type Store interface {
	Get(ctx context.Context, uid string) (Datasource, error)
	Insert(ctx context.Context, ds Datasource) error
	Update(ctx context.Context, ds Datasource) error
	Delete(ctx context.Context, uid string) error
}

// Coremodel
type Coremodel struct {
	store  Store
	client client.Client
	schema schema.ObjectSchema
}

// ProvideCoremodel
func ProvideCoremodel(store Store, schemaLib thema.Library) (*Coremodel, error) {
	schema, err := schema.LoadThemaSchema(
		cuePath,
		cueFS,
		schemaLib,
		schemaVersion,
		&DatasourceSpec{},
		groupName,
		groupVersion,
		schemaOpenapi,
		&Datasource{},
		&DatasourceList{},
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
func (m *Coremodel) RegisterController(mgr ctrl.Manager) error {
	m.client = mgr.GetClient()

	return ctrl.NewControllerManagedBy(mgr).
		For(&Datasource{}).
		Complete(m)
}

// Reconcile
func (m *Coremodel) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	fmt.Println("Got for reconciliation", req.String())

	ds := Datasource{}
	err := m.client.Get(ctx, req.NamespacedName, &ds)

	if kerrors.IsNotFound(err) {
		fmt.Println("Resource not found in k8s, deleting from store")

		if err := m.store.Delete(ctx, req.Name); err != nil {
			fmt.Println("Error deleting from store:", err)

			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: 1 * time.Minute,
			}, err
		}

		fmt.Println("OK deleted resource from store")

		return reconcile.Result{}, nil
	}

	if err != nil {
		fmt.Println("Error fetching resource from k8s:", err)

		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 1 * time.Minute,
		}, err
	}

	_, err = m.store.Get(ctx, string(ds.UID))
	if err != nil {
		if !errors.Is(err, models.ErrDataSourceNotFound) {
			fmt.Println("Error getting resource from the store:", err)

			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: 1 * time.Minute,
			}, err
		}

		fmt.Println("Resource not found in store, inserting from k8s")
		if err := m.store.Insert(ctx, ds); err != nil {
			fmt.Println("Error inserting resource into store:", err)
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: 1 * time.Minute,
			}, err
		}

		fmt.Println("OK inserted resource into store")

		return reconcile.Result{}, nil
	}

	fmt.Println("Resource is found in store, updating from k8s")
	if err := m.store.Update(ctx, ds); err != nil {
		fmt.Println("Error updating resource in store:", err)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 1 * time.Minute,
		}, err
	}

	fmt.Println("OK updated resource in store")
	return reconcile.Result{}, nil
}
