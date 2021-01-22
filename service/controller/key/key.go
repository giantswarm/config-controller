package key

import (
	"regexp"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/config-controller/pkg/generator/key"
)

// LegacyConfigVersion should be set when the config for the app should not
// be generated.
const LegacyConfigVersion = "0.0.0"

const Owner = key.Owner

var (
	tagConfigVersionPattern = regexp.MustCompile(`^(\d+)\.x\.x$`)
)

type Object interface {
	runtime.Object

	GetAnnotations() map[string]string
	GetName() string
	GetNamespace() string
	SetAnnotations(map[string]string)
	GetObjectKind() schema.ObjectKind
}

func GetObjectHash(o Object) (string, bool) {
	return key.GetObjectHash(o)
}

func GetObjectKind(o Object) string {
	return o.GetObjectKind().GroupVersionKind().Kind
}

func NamespacedName(o Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: o.GetNamespace(),
		Name:      o.GetName(),
	}
}

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

// TryVersionToTag translates config version: "<major>.x.x" to tagReference:
// "v<major>" if possible. Otherwise returns empty string.
func TryVersionToTag(version string) string {
	matches := tagConfigVersionPattern.FindAllStringSubmatch(version, -1)
	if len(matches) > 0 {
		return "v" + matches[0][1]
	}
	return ""
}

func RemoveAnnotation(annotations map[string]string, key string) map[string]string {
	if annotations == nil {
		return nil
	}

	_, ok := annotations[key]
	if ok {
		delete(annotations, key)
	}

	return annotations
}
