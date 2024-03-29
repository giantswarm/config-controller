package k8sresource

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Config struct {
	Client k8sclient.Interface
	Logger micrologger.Logger
}

type Service struct {
	logger micrologger.Logger

	client client.Client
	scheme *runtime.Scheme
}

func New(config Config) (*Service, error) {
	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	s := &Service{
		logger: config.Logger,

		client: config.Client.CtrlClient(),
		scheme: config.Client.Scheme(),
	}

	return s, nil
}

func (s *Service) EnsureCreated(ctx context.Context, hashAnnotation string, desired client.Object) error {
	s.logger.Debugf(ctx, "ensuring %#q %#q", s.Kind(desired), ObjectKey(desired))

	err := setHash(hashAnnotation, desired)
	if err != nil {
		return microerror.Mask(err)
	}

	t := reflect.TypeOf(desired).Elem()
	current := reflect.New(t).Interface().(client.Object)
	err = s.client.Get(ctx, ObjectKey(desired), current)
	if apierrors.IsNotFound(err) {
		err = s.client.Create(ctx, desired)
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.Debugf(ctx, "created %#q %#q", s.Kind(desired), ObjectKey(desired))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	h1, ok1 := GetAnnotation(desired, hashAnnotation)
	h2, ok2 := GetAnnotation(current, hashAnnotation)

	if ok1 && ok2 && h1 == h2 {
		s.logger.Debugf(ctx, "object %#q %#q is up to date", s.Kind(desired), ObjectKey(desired))
		return nil
	}

	err = s.client.Update(ctx, desired)
	if err != nil {
		return microerror.Mask(err)
	}

	s.logger.Debugf(ctx, "updated %#q %#q", s.Kind(desired), ObjectKey(desired))
	return nil
}

func (s *Service) EnsureDeleted(ctx context.Context, obj client.Object) error {
	s.logger.Debugf(ctx, "ensuring deletion of %#q %#q", s.Kind(obj), ObjectKey(obj))

	err := s.client.Delete(ctx, obj)
	if apierrors.IsNotFound(err) {
		s.logger.Debugf(ctx, "%#q %#q already deleted", s.Kind(obj), ObjectKey(obj))
	} else if err != nil {
		return microerror.Mask(err)
	}

	s.logger.Debugf(ctx, "deleted %#q %#q", s.Kind(obj), ObjectKey(obj))

	return nil
}

func (s *Service) GroupVersionKind(o Object) (schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(o, s.scheme)
	if err != nil {
		return schema.GroupVersionKind{}, microerror.Mask(err)
	}

	return gvk, nil
}

// Modify gets the object for the given key. It sets the most recent version of
// the object to provided obj pointer and calls modifyFunc which is supposed to
// apply changes to the pointer.
//
//   - The modifyFunc is called on every try.
//   - The obj variable is reset and populated before every try.
//   - There are no retries if the object defined by the key does not exist.
//
// Example usage:
//
//	key := client.ObjectKey{Namespace: "giantswarm", Name: "my-operator"}
//	current := &v1alpha1.App{}
//	modifyFunc := func() error {
//		current.Spec.Version = "2.0.0"
//		return nil
//	}
//	err := h.resource.Modify(ctx, key, current, modifyFunc, nil)
//	if err != nil {
//		...
//	}
func (s *Service) Modify(ctx context.Context, key client.ObjectKey, obj client.Object, modifyFunc func() error, backOff backoff.BackOff) error {
	err := s.modify(ctx, key, obj, modifyFunc, backOff, false)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// ModifyStatus works exactly like Modify but updates the status subresource.
func (s *Service) ModifyStatus(ctx context.Context, key client.ObjectKey, obj client.Object, modifyFunc func() error, backOff backoff.BackOff) error {
	err := s.modify(ctx, key, obj, modifyFunc, backOff, true)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (s *Service) modify(ctx context.Context, key client.ObjectKey, obj client.Object, modifyFunc func() error, backOff backoff.BackOff, statusUpdate bool) error {
	if obj == nil {
		panic("nil obj")
	}

	if statusUpdate {
		s.logger.Debugf(ctx, "modifying status %#q %#q", s.Kind(obj), key)
	} else {
		s.logger.Debugf(ctx, "modifying %#q %#q", s.Kind(obj), key)
	}

	v := reflect.ValueOf(obj)

	// Make sure we have a pointer behind the interface.
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("value of zero.(%s) of kind %q expected to be %q", v.Type(), v.Kind(), reflect.Ptr))
	}

	// Make sure the pointer has a value set.
	if v.IsZero() {
		panic(fmt.Sprintf("value behind obj.(%s) pointer is nil (%v)", v.Type(), obj))
	}

	if backOff == nil {
		backOff = backoff.NewMaxRetries(6, 150*time.Millisecond)
	}

	attempt := 0

	o := func() error {
		var err error

		attempt++

		// Zero the value behind the pointer.
		e := v.Elem()
		e.Set(reflect.Zero(e.Type()))

		err = s.client.Get(ctx, key, obj)
		if apierrors.IsNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}

		err = modifyFunc()
		if err != nil {
			return microerror.Mask(err)
		}

		if statusUpdate {
			err = s.client.Status().Update(ctx, obj)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			err = s.client.Update(ctx, obj)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}
	n := func(err error, d time.Duration) {
		if statusUpdate {
			s.logger.Debugf(ctx, "retrying (%d) %#q %#q status modification in %s due to error: %s", attempt, s.Kind(obj), ObjectKey(obj), d, err)
		} else {
			s.logger.Debugf(ctx, "retrying (%d) %#q %#q modification in %s due to error: %s", attempt, s.Kind(obj), ObjectKey(obj), d, err)
		}
	}
	err := backoff.RetryNotify(o, backOff, n)
	if err != nil {
		return microerror.Mask(err)
	}

	if statusUpdate {
		s.logger.Debugf(ctx, "modified status %#q %#q", s.Kind(obj), key)
	} else {
		s.logger.Debugf(ctx, "modified %#q %#q", s.Kind(obj), key)
	}

	return nil
}

// Kind is a best effort approach to extract object kind. It should serve only
// logging purposes.
func (s *Service) Kind(o Object) string {
	gvk, err := apiutil.GVKForObject(o, s.scheme)
	if err != nil {
		t := fmt.Sprintf("%T", o)
		t = t[strings.LastIndex(t, ".")+1:]
		return t
	}

	return gvk.Kind
}

func setHash(annotation string, o Object) error {
	bytes, err := json.Marshal(o)
	if err != nil {
		return microerror.Mask(err)
	}

	sum := sha256.Sum256(bytes)
	SetAnnotation(o, annotation, fmt.Sprintf("%x", sum))

	return nil
}
