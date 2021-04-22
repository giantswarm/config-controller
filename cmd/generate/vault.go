package generate

import (
	"context"
	"errors"
	"os"

	"github.com/giantswarm/microerror"
	vaultapi "github.com/hashicorp/vault/api"
)

type vaultClientConfig struct {
	Address string `json:"addr"`
	Token   string `json:"token"`
	CAPath  string `json:"caPath"`
}

func newVaultClient(config vaultClientConfig) (*vaultapi.Client, error) {
	c := vaultapi.DefaultConfig()
	c.Address = config.Address
	c.MaxRetries = 4 // Total of 5 tries.
	err := c.ConfigureTLS(&vaultapi.TLSConfig{
		CAPath: config.CAPath,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	vaultClient, err := vaultapi.NewClient(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	vaultClient.SetToken(config.Token)

	return vaultClient, nil
}

func createVaultClientUsingEnv(ctx context.Context, installation string) (*vaultapi.Client, error) {
	if os.Getenv("VAULT_ADDR") == "" {
		return nil, microerror.Mask(errors.New("VAULT_ADDR must beset"))
	}
	config := vaultClientConfig{
		Address: os.Getenv("VAULT_ADDR"),
		Token:   os.Getenv("VAULT_TOKEN"),
		CAPath:  os.Getenv("VAULT_CAPATH"),
	}

	vaultClient, err := newVaultClient(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return vaultClient, nil

}
