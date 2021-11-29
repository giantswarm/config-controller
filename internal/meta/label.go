package meta

import (
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/config-controller/pkg/project"
)

var (
	managedByLabel = label.ManagedBy
	versionLabel   = label.ConfigControllerVersion
)

type ManagedBy struct{}

func (ManagedBy) Key() string { return managedByLabel }

func (ManagedBy) Default() string { return project.Name() }

type Version struct{}

func (Version) Key() string { return versionLabel }

func (Version) Val(uniqueApp bool) string {
	if uniqueApp {
		// When config-controller is deployed as a unique app it only
		// processes management cluster CRs. These CRs always have the
		// version label config-controller.giantswarm.io/version: 0.0.0
		return "0.0.0"
	} else {
		return project.Version()
	}
}

func (Version) Selector(uniqueApp bool) (labels.Selector, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", label.ConfigControllerVersion, (Version{}).Val(uniqueApp)))
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return selector, nil
}
