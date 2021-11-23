module github.com/giantswarm/config-controller

go 1.14

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/fatih/color v1.13.0
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions-application v0.1.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.1
	github.com/giantswarm/k8sclient/v6 v6.0.0
	github.com/giantswarm/k8smetadata v0.6.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v6 v6.0.0
	github.com/giantswarm/valuemodifier v0.3.1
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-test/deep v1.0.7 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-retryablehttp v0.6.7 // indirect
	github.com/hashicorp/vault/api v1.3.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	k8s.io/api v0.20.12
	k8s.io/apimachinery v0.20.12
	k8s.io/client-go v0.20.12
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.3.0
)
