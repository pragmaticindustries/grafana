package k8sbridge

import (
	"context"
	"fmt"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/grafana/grafana/pkg/schema"
)

// Clientset
type Clientset struct {
	cfg *rest.Config
	// TODO: this needs to be exposed, but only specific types (e.g. no pods / deployments / etc.).
	coreset *kubernetes.Clientset
	extset  *apiextensionsclient.Clientset
	crds    map[k8schema.GroupVersion]apiextensionsv1.CustomResourceDefinition
}

// NewClientset
func NewClientset(cfg *rest.Config, schemas schema.CoreSchemaList) (*Clientset, error) {
	k8sset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	extset, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Clientset{
		coreset: k8sset,
		extset:  extset,
		crds:    make(map[k8schema.GroupVersion]apiextensionsv1.CustomResourceDefinition),
	}, nil
}

// RegisterSchema
func (c *Clientset) RegisterSchema(ctx context.Context, s schema.ObjectSchema) error {
	ver := k8schema.GroupVersion{
		Group:   s.GroupName(),
		Version: s.GroupVersion(),
	}

	crdObj := newCRD(s.Name(), s.GroupName(), s.GroupVersion(), s.OpenAPISchema())
	crd, err := c.extset.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Create(ctx, &crdObj, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	c.crds[ver] = *crd

	return nil
}

func newCRD(
	objectKind, groupName, groupVersion string, schema apiextensionsv1.JSONSchemaProps,
) apiextensionsv1.CustomResourceDefinition {
	return apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%ss.%s", objectKind, groupName),
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: groupName,
			Scope: apiextensionsv1.NamespaceScoped, // TODO: make configurable?
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   objectKind + "s", // TODO: figure out better approach?
				Singular: objectKind,
				Kind:     capitalize(objectKind),
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    groupVersion,
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &schema,
					},
				},
			},
		},
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	u := strings.ToUpper(string(s[0]))

	if len(s) == 1 {
		return u
	}

	return u + s[1:]
}
