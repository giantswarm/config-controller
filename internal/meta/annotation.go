package meta

import (
	"encoding/json"
	"os"
	"os/user"

	apiextensionsannotation "github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/project"
)

var (
	configVersionAnnotation   = apiextensionsannotation.ConfigVersion
	xAppAnnotation            = project.Name() + ".x-giantswarm.io/app"
	xCreatorAnnotation        = project.Name() + ".x-giantswarm.io/creator"
	xInstallationAnnotation   = project.Name() + ".x-giantswarm.io/installation"
	xObjectHashAnnotation     = project.Name() + ".x-giantswarm.io/object-hash"
	xPreviousConfigAnnotation = project.Name() + ".x-giantswarm.io/previous-config"
	xProjectVersionAnnotation = project.Name() + ".x-giantswarm.io/project-version"
)

type ConfigVersion struct{}

func (ConfigVersion) Key() string { return configVersionAnnotation }

type XPreviousConfig struct{}

func (XPreviousConfig) Key() string { return xPreviousConfigAnnotation }

func (XPreviousConfig) Get(o Object) (v1alpha1.ConfigStatusConfig, bool, error) {
	a := o.GetAnnotations()
	if a == nil {
		return v1alpha1.ConfigStatusConfig{}, false, nil
	}

	raw, ok := a[xPreviousConfigAnnotation]
	if !ok {
		return v1alpha1.ConfigStatusConfig{}, false, nil
	}

	var c v1alpha1.ConfigStatusConfig
	err := json.Unmarshal([]byte(raw), &c)
	if err != nil {
		return v1alpha1.ConfigStatusConfig{}, false, microerror.Mask(err)
	}

	return c, true, nil
}

func (XPreviousConfig) Set(o Object, c v1alpha1.ConfigStatusConfig) error {
	bs, err := json.Marshal(c)
	if err != nil {
		return microerror.Mask(err)
	}

	a := o.GetAnnotations()
	if a == nil {
		a = make(map[string]string, 1)
	}

	a[xPreviousConfigAnnotation] = string(bs)

	o.SetAnnotations(a)
	return nil
}

type XApp struct{}

func (XApp) Key() string { return xAppAnnotation }

type XCreator struct{}

func (XCreator) Key() string { return xCreatorAnnotation }

func (XCreator) Default() string {
	u, err := user.Current()
	if err != nil {
		return u.Username
	}

	if os.Getenv("USER") != "" {
		return os.Getenv("USER")
	}

	return os.Getenv("USERNAME") // Windows
}

type XInstallation struct{}

func (XInstallation) Key() string { return xInstallationAnnotation }

type XObjectHash struct{}

func (XObjectHash) Key() string { return xObjectHashAnnotation }

type XProjectVersion struct{}

func (XProjectVersion) Key() string { return xProjectVersionAnnotation }

func (XProjectVersion) Val(unique bool) string {
	if unique {
		return "0.0.0"
	}

	return project.Version()
}
