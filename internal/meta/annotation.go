package meta

import (
	"os"
	"os/user"

	apiextensionsannotation "github.com/giantswarm/apiextensions/v3/pkg/annotation"

	"github.com/giantswarm/config-controller/pkg/project"
)

var (
	configVersionAnnotation   = apiextensionsannotation.ConfigVersion
	xAppAnnotation            = project.Name() + ".x-giantswarm.io/app"
	xCreatorAnnotation        = project.Name() + ".x-giantswarm.io/creator"
	xInstallationAnnotation   = project.Name() + ".x-giantswarm.io/installation"
	xProjectVersionAnnotation = project.Name() + ".x-giantswarm.io/project-version"
)

func (annotation) ConfigVersion() string {
	return configVersionAnnotation
}

func (annotation) XApp() string {
	return xAppAnnotation
}

func (annotation) XCreator() string {
	return xCreatorAnnotation
}

func (annotation) XCreatorDefault() string {
	u, err := user.Current()
	if err != nil {
		return u.Username
	}

	if os.Getenv("USER") != "" {
		return os.Getenv("USER")
	}

	return os.Getenv("USERNAME") // Windows
}

func (annotation) XInstallation() string {
	return xInstallationAnnotation
}

func (annotation) XProjectVersion() string {
	return xProjectVersionAnnotation
}

func (annotation) XProjectVersionVal(unique bool) string {
	if unique {
		return "0.0.0"
	}

	return project.Version()
}
