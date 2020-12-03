package values

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/config-controller/pkg/generator"
	controllerkey "github.com/giantswarm/config-controller/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	app, err := controllerkey.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	configVersion, ok := app.GetAnnotations()[annotation.ConfigVersion]
	if !ok {
		r.logger.Debugf(ctx, "App CR %q is missing %q annotation", app.Name, annotation.ConfigVersion)
		r.logger.Debugf(ctx, "cancelling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Spec.Config.ConfigMap.Name,
			Namespace: app.Spec.Config.ConfigMap.Namespace,
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Spec.Config.Secret.Name,
			Namespace: app.Spec.Config.Secret.Namespace,
		},
	}
	if cm.Name == "" || secret.Name == "" {
		ref := configVersion
		if tagRef := controllerkey.TryVersionToTag(configVersion); tagRef != "" {
			ref = tagRef
		}
		name := generator.GenerateResourceName(app.Spec.Name, ref)
		if cm.Name == "" {
			cm.Name = name
			cm.Namespace = app.Namespace
		}
		if secret.Name == "" {
			secret.Name = name
			secret.Namespace = app.Namespace
		}
	}

	r.logger.Debugf(ctx, "deleting App %#q, config version %#q", app.Spec.Name, configVersion)

	r.logger.Debugf(ctx, "clearing App %#q, config version %#q configmap and secret details", app.Spec.Name, configVersion)
	app.Spec.Config = v1alpha1.AppSpecConfig{}
	err = r.k8sClient.CtrlClient().Update(ctx, &app)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "cleared App %#q, config version %#q configmap and secret details", app.Spec.Name, configVersion)

	r.logger.Debugf(ctx, "deleting configmap for App %#q, config version %#q", app.Spec.Name, configVersion)
	err = r.k8sClient.CtrlClient().Delete(ctx, cm)
	if client.IgnoreNotFound(err) != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "deleted configmap for App %#q, config version %#q", app.Spec.Name, configVersion)

	r.logger.Debugf(ctx, "deleting secret for App %#q, config version %#q", app.Spec.Name, configVersion)
	err = r.k8sClient.CtrlClient().Delete(ctx, secret)
	if client.IgnoreNotFound(err) != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "deleted secret for App %#q, config version %#q", app.Spec.Name, configVersion)

	r.logger.Debugf(ctx, "deleted App %#q, config version %#q", app.Spec.Name, configVersion)

	return nil
}
