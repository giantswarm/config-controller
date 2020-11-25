package key

import (
	"github.com/giantswarm/config-controller/pkg/project"
)

const (
	Owner          = "giantswarm"
	ManagedByLabel = "giantswarm.io/managed-by"
)

var (
	ConfigVersionAnnotation = project.Name() + ".giantswarm.io/config-version"
)
