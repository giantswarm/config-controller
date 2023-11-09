package key

import (
	"regexp"

	"github.com/giantswarm/microerror"

	corev1alpha1 "github.com/giantswarm/config-controller/api/v1alpha1"
)

const (
	// LegacyConfigVersion should be set when the config for the app should not
	// be generated.
	LegacyConfigVersion = "0.0.0"

	ObjectHashAnnotation = "config-controller.giantswarm.io/object-hash"
)

var (
	tagConfigVersionPattern = regexp.MustCompile(`^(\d+)\.x\.x$`)
)

func ToConfigCR(v interface{}) (*corev1alpha1.Config, error) {
	if v == nil {
		return nil, microerror.Maskf(wrongTypeError, "expected non-nil, got %#v", v)
	}

	p, ok := v.(*corev1alpha1.Config)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected %T, got %T", p, v)
	}

	return p.DeepCopy(), nil
}
