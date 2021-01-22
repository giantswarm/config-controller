package key

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/config-controller/pkg/project"
)

const (
	Owner                      = "giantswarm"
	KubernetesManagedByLabel   = "app.kubernetes.io/managed-by"
	ReleaseNameAnnotation      = "meta.helm.sh/release-name"
	ReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
)

type object interface {
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
}

var objectHashAnnotation = project.Name() + ".giantswarm.io/object-hash"

func GetObjectHash(o object) (string, bool) {
	a := o.GetAnnotations()
	if len(a) == 0 {
		return "", false
	}

	sum := a[objectHashAnnotation]
	return sum, sum != ""
}

func SetObjectHash(o object) {
	bytes, err := json.Marshal(o)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal to JSON: %+v with error: %s\n", o, err))
	}
	sum := sha256.Sum256(bytes)

	a := o.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a[objectHashAnnotation] = fmt.Sprintf("%x", sum)
}
