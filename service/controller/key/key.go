package key

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
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
