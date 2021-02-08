package label

import (
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/config-controller/pkg/project"
)

func AppVersionSelector(unique bool) (labels.Selector, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", label.ConfigControllerVersion, GetProjectVersion(unique)))
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return selector, nil
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
