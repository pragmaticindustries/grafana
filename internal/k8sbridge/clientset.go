package k8sbridge

import (
	"context"
	"fmt"

	"cuelang.org/go/pkg/strings"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/grafana/grafana/pkg/schema"
)

// Clientset
type Clientset struct {
	k8sset  *kubernetes.Clientset
	extset  *apiextensionsclient.Clientset
	coreset map[k8schema.GroupVersion]*rest.RESTClient
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

	coreset := make(map[k8schema.GroupVersion]*rest.RESTClient, len(schemas))
	crds := make(map[k8schema.GroupVersion]apiextensionsv1.CustomResourceDefinition, len(schemas))
	for _, s := range schemas {
		ver := k8schema.GroupVersion{
			Group:   s.GroupName(),
			Version: s.GroupVersion(),
		}

		c := *cfg
		c.NegotiatedSerializer = clientscheme.Codecs.WithoutConversion()
		c.GroupVersion = &ver

		cli, err := rest.RESTClientFor(&c)
		if err != nil {
			return nil, err
		}

		crdObj := NewCRD(s.Name(), s.GroupName(), s.GroupVersion(), s.OpenAPISchema())
		crd, err := extset.
			ApiextensionsV1().
			CustomResourceDefinitions().
			Create(
				context.TODO(),
				&crdObj,
				metav1.CreateOptions{},
			)
		if err != nil && !errors.IsAlreadyExists(err) {
			return nil, err
		}

		crds[ver] = *crd
		coreset[ver] = cli
	}

	return &Clientset{
		k8sset:  k8sset,
		extset:  extset,
		coreset: coreset,
		crds:    crds,
	}, nil
}

// ClientForVersion
func (c *Clientset) ClientForSchema(schema schema.ObjectSchema) (*rest.RESTClient, error) {
	k := k8schema.GroupVersion{
		Group:   schema.GroupName(),
		Version: schema.GroupVersion(),
	}

	v, ok := c.coreset[k]
	if !ok {
		return nil, fmt.Errorf("no client registered for schema: %s/%s", schema.GroupName(), schema.GroupVersion())
	}

	return v, nil
}

// NewCRD
// TODO: use these to automatically register CRDs to the server.
func NewCRD(
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
