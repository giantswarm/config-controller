package key

import (
	"github.com/giantswarm/config-controller/pkg/project"
)

const (
	Owner                      = "giantswarm"
	KubernetesManagedByLabel   = "app.kubernetes.io/managed-by"
	GiantswarmManagedByLabel   = "giantswarm.io/managed-by"
	ReleaseNameAnnotation      = "meta.helm.sh/release-name"
	ReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
)

var (
	ConfigVersion = project.Name() + ".giantswarm.io/config-version"
	AppLabel      = project.Name() + ".giantswarm.io/app"
)

func ToAppCR(v interface{}) (v1alpha1.App, error) {
	if v == nil {
		return v1alpha1.App{}, microerror.Maskf(wrongTypeError, "expected non-nil, got %#v", v)
	}

	p, ok := v.(*v1alpha1.App)
	if !ok {
		return v1alpha1.App{}, microerror.Maskf(wrongTypeError, "expected %T, got %T", p, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
