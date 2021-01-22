package k8sresource

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	Client client.Client
	Logger micrologger.Logger
}

type Service struct {
	client client.Client
	logger micrologger.Logger
}

func New(config Config) (*Service, error) {
	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	s := &Service{
		client: config.Client,
		logger: config.Logger,
	}

	return s, nil
}

func (s *Service) EnsureCreated(ctx context.Context, hashAnnotation string, desired Object) error {
	s.logger.Debugf(ctx, "ensuring %#q %#q", Kind(desired), ObjectKey(desired))

	err := setHash(hashAnnotation, desired)
	if err != nil {
		return microerror.Mask(err)
	}

	t := reflect.TypeOf(desired).Elem()
	current := reflect.New(t).Interface().(Object)
	err = s.client.Get(ctx, ObjectKey(desired), current)
	if apierrors.IsNotFound(err) {
		err = s.client.Create(ctx, desired)
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.Debugf(ctx, "created %#q %#q", Kind(desired), ObjectKey(desired))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	h1, ok1 := GetAnnotation(desired, hashAnnotation)
	h2, ok2 := GetAnnotation(current, hashAnnotation)

	if ok1 && ok2 && h1 == h2 {
		s.logger.Debugf(ctx, "object %#q %#q is up to date", Kind(desired), ObjectKey(desired))
		return nil
	}

	err = s.client.Update(ctx, desired)
	if err != nil {
		return microerror.Mask(err)
	}

	s.logger.Debugf(ctx, "updated %#q %#q", Kind(desired), ObjectKey(desired))
	return nil
}

func (s *Service) Modify(ctx context.Context, key client.ObjectKey, obj Object, modifyFunc func() error, backOff backoff.BackOff) error {
	if obj == nil {
		panic("nil obj")
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
		backOff = backoff.NewMaxRetries(3, 300*time.Millisecond)
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

		err = s.client.Update(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	n := func(err error, d time.Duration) {
		s.logger.Debugf(ctx, "retrying (%d) %#q %#q modification in %s due to error: %s", attempt, Kind(obj), ObjectKey(obj), d, err)
	}
	err := backoff.RetryNotify(o, backOff, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
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
