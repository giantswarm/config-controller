package key

import (
	"github.com/giantswarm/config-controller/pkg/project"
)

var (
	ConfigVersionAnnotation = project.Name() + ".giantswarm.io/config-version"
)
