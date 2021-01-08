package configversion

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/service/controller/key"
)

func (h *Handler) EnsureCreated(ctx context.Context, obj interface{}) error {
	app, err := key.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	annotations := app.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if app.Spec.Catalog == "" {
		h.logger.Debugf(ctx, "App CR has no .Spec.Catalog set")
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	if app.Spec.Catalog == "releases" {
		h.logger.Debugf(ctx, "App CR has a \"releases\" catalog set")
		h.logger.Debugf(ctx, "removing %#q annotation", key.PauseAnnotation)
		app.SetAnnotations(key.RemoveAnnotation(annotations, key.PauseAnnotation))
		err = h.k8sClient.CtrlClient().Update(ctx, &app)
		if err != nil {
			return microerror.Mask(err)
		}
		h.logger.Debugf(ctx, "removed %#q annotation", key.PauseAnnotation)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	h.logger.Debugf(ctx, "setting App config version")

	h.logger.Debugf(ctx, "resolving config version from %#q catalog", app.Spec.Catalog)
	var index Index
	{
		indexYamlBytes, err := getCatalogIndex(ctx, app.Spec.Catalog)
		if err != nil {
			return microerror.Mask(err)
		}

		err = yaml.Unmarshal(indexYamlBytes, &index)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	entries, ok := index.Entries[app.Spec.Name]
	if !ok || len(entries) == 0 {
		h.logger.Debugf(ctx, "App has no entries in %#q's index.yaml", app.Spec.Catalog)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	var configVersion string
	for _, entry := range entries {
		if entry.Version == app.Spec.Version {
			if entry.ConfigVersion != "" {
				configVersion = entry.ConfigVersion
			} else {
				configVersion = "0.0.0"
			}
			break
		}
	}

	if configVersion == "" {
		h.logger.Debugf(ctx, "App has no entries matching version %#q in %#q's index.yaml", app.Spec.Version, app.Spec.Catalog)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}
	h.logger.Debugf(ctx, "resolved config version from %#q catalog to %#q", app.Spec.Catalog, configVersion)

	if v, ok := annotations[annotation.ConfigVersion]; ok {
		if v == configVersion {
			h.logger.Debugf(ctx, "App has correct version annotation already")
			h.logger.Debugf(ctx, "cancelling handler")
			return nil
		}
	}

	annotations[annotation.ConfigVersion] = configVersion
	app.SetAnnotations(annotations)
	err = h.k8sClient.CtrlClient().Update(ctx, &app)
	if err != nil {
		return microerror.Mask(err)
	}
	h.logger.Debugf(ctx, "set config version to %#q", configVersion)

	return nil
}

func getCatalogIndex(ctx context.Context, catalog string) ([]byte, error) {
	url := fmt.Sprintf("https://giantswarm.github.io/%s/index.yaml", catalog)
	response, err := http.Get(url)
	if err != nil {
		return []byte{}, microerror.Mask(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, microerror.Mask(err)
	}

	return body, nil
}
