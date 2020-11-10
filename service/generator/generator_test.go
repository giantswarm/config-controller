package generator

import (
	"testing"
)

func TestGenerator_GenerateConfig(t *testing.T) {
	testCases := []struct {
		name string

		fs           Filesystem
		app          string
		installation string

		expectedConfigmap string
		expectedSecret    string
	}{
		{
			name: "case 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				Fs: tc.fs,
			}

			g, err := New(&config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			configmap, secret, err := g.GenerateConfig(tc.installation, tc.app)
			if configmap != tc.expectedConfigmap {
				t.Fatalf("configmap not expected, got: %s", configmap)
			}
			if secret != tc.expectedSecret {
				t.Fatalf("secret not expected, got: %s", secret)
			}
		})
	}
}
