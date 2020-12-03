package label

import (
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/operatorkit/v4/pkg/controller"

	"github.com/giantswarm/config-controller/pkg/project"
)

func AppVersionSelector(unique bool) controller.Selector {
	return controller.NewSelector(func(labels controller.Labels) bool {
		if !labels.Has(label.ConfigControllerVersion) {
			return false
		}
		if labels.Get(label.ConfigControllerVersion) == GetProjectVersion(unique) {
			return true
		}

		return false
	})
}

func GetProjectVersion(unique bool) string {
	if unique {
		// When config-controller is deployed as a unique app it only
		// processes control plane app CRs. These CRs always have the
		// version label config-controller.giantswarm.io/version: 0.0.0
		return project.AppControlPlaneVersion()
	} else {
		return project.Version()
	}
}
