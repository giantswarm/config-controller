package generator

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/giantswarm/microerror"
)

func TestGenerator_GenerateConfig(t *testing.T) {
	testCases := []struct {
		name string

		app          string
		installation string

		// mandatory
		configYaml         string
		configmapTemplate  string
		installationSecret string
		secretTemplate     string
		// optional
		configYamlPatch        string
		configmapTemplatePatch string
		includeFiles           map[string]string

		expectedConfigmap string
		expectedSecret    string
	}{
		{
			name:         "case 0 - basic templates and installation-level config.yaml patch",
			app:          "operator",
			installation: "puma",

			configYaml: "universalValue: 42",
			configmapTemplate: `
answer: {{ .universalValue }}
region: {{ .provider.region }}`,
			installationSecret: "key: password",
			secretTemplate:     `secretAccessKey: {{ .key }}`,

			configYamlPatch: "provider: {kind: aws, region: us-east-1}",

			expectedConfigmap: `
answer: 42
region: us-east-1`,
			expectedSecret: `secretAccessKey: password`,
		},

		{
			name:         "case 1 - include files from include/ directory",
			app:          "operator",
			installation: "puma",

			configYaml: "universalValue: 42",
			configmapTemplate: `
answer: {{ .universalValue }}
region: {{ .provider.region }}
availableInstances:
{{- include "instances" . | nindent 2 }}
`,
			installationSecret: "key: password",
			secretTemplate:     `secretAccessKey: {{ .key }}`,

			configYamlPatch: "provider: {kind: aws, region: us-east-1}",
			includeFiles: map[string]string{
				"instances.yaml": "- small\n- medium\n- large",
			},

			expectedConfigmap: `
answer: 42
availableInstances:
- small
- medium
- large
region: us-east-1`,
			expectedSecret: `secretAccessKey: password`,
		},

		{
			name:         "case 2 - overriding default app template with installation-specific app template",
			app:          "operator",
			installation: "puma",

			configYaml: `
universalValue: 42
registry: docker.io`,
			configmapTemplate: `
answer: {{ .universalValue }}
region: {{ .provider.region }}
registry: {{ .registry }}`,
			installationSecret: "key: password",
			secretTemplate:     `secretAccessKey: {{ .key }}`,

			configYamlPatch:        "provider: {kind: aws, region: us-east-1}",
			configmapTemplatePatch: "registry: azurecr.io",

			expectedConfigmap: `
answer: 42
region: us-east-1
registry: azurecr.io`,
			expectedSecret: `secretAccessKey: password`,
		},

		{
			name:         "case 3 - template integers",
			app:          "operator",
			installation: "puma",

			configYaml: "universalValue: 42",
			configmapTemplate: `
answer: {{ .universalValue }}
region: {{ .provider.region }}`,
			installationSecret: "key: 123456",
			secretTemplate:     `secretAccessKey: {{ .key }}`,

			configYamlPatch: "provider: {kind: aws, region: us-east-1}",
			configmapTemplatePatch: `answer: 5
exampleInt: 33
exampleFloat: 13.2
`,

			expectedConfigmap: `
answer: 5
exampleFloat: 13.2
exampleInt: 33
region: us-east-1
`,
			expectedSecret: `secretAccessKey: 123456`,
		},

		{
			name:         "case 4 - templating in included files",
			app:          "operator",
			installation: "puma",

			configYaml: "universalValue: 42\nextraValue: 43",
			configmapTemplate: `
answer: {{ .universalValue }}
{{ include "templated-include" . }}
`,
			installationSecret: "key: 123456",
			secretTemplate:     `secretAccessKey: {{ .key }}`,

			includeFiles: map[string]string{
				"templated-include.yaml": "exampleObj: {{ .extraValue }}",
			},
			expectedConfigmap: `
answer: 42
exampleObj: 43`,
			expectedSecret: `secretAccessKey: 123456`,
		},

		{
			name:         "case 5 - complex indent with include",
			app:          "operator",
			installation: "puma",

			configYaml: "universalValue: 42",
			configmapTemplate: `
answer: {{ .universalValue }}
level1:
  {{- include "level1" . | nindent 2 }}
  level2:
    {{- include "level2" . | nindent 4 }}
    level3:
      {{- include "level3" . | nindent 6 }}
    {{- include "level2-2" . | nindent 4 }}
  {{- include "level1-2" . | nindent 2 }}
`,
			installationSecret: "key: 123456",
			secretTemplate:     `secretAccessKey: {{ .key }}`,

			includeFiles: map[string]string{
				"level1.yaml":   "firstLevel: true",
				"level2.yaml":   "secondLevel: true",
				"level3.yaml":   "thirdLevel: true",
				"level2-2.yaml": "backOnSecond: true",
				"level1-2.yaml": "backOnFirst: true",
			},
			expectedConfigmap: `
answer: 42
level1:
  backOnFirst: true
  firstLevel: true
  level2:
    backOnSecond: true
    level3:
      thirdLevel: true
    secondLevel: true
`,
			expectedSecret: `secretAccessKey: 123456`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := newMockFilesystem(
				tc.installation, tc.app,
				sanitize(tc.configYaml),
				sanitize(tc.configmapTemplate),
				sanitize(tc.installationSecret),
				sanitize(tc.secretTemplate),
			)
			if tc.configYamlPatch != "" {
				fs.AddConfigPatch(sanitize(tc.configYamlPatch))
			}
			if tc.configmapTemplatePatch != "" {
				fs.AddConfigmapTemplatePatch(sanitize(tc.configmapTemplatePatch))
			}
			for name, contents := range tc.includeFiles {
				fs.AddIncludeFile(name, sanitize(contents))
			}

			config := Config{
				Fs: fs,
			}

			g, err := New(&config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			configmap, secret, err := g.GenerateConfig(tc.installation, tc.app)
			if err != nil {
				t.Fatalf("unexpected error: %s", microerror.Pretty(err, true))
			}
			if sanitize(configmap) != sanitize(tc.expectedConfigmap) {
				t.Fatalf("configmap not expected, got: %s", configmap)
			}
			if sanitize(secret) != sanitize(tc.expectedSecret) {
				t.Fatalf("secret not expected, got: %s", secret)
			}
		})
	}
}

func sanitize(in string) string {
	return strings.TrimSpace(
		strings.ReplaceAll(in, "\t", ""),
	)
}

type mockFilesystem struct {
	app          string
	installation string

	files map[string][]byte
}

func newMockFilesystem(installation, app, globalConfig, appConfigmapTemplate, installationSecret, appSecretTemplate string) *mockFilesystem {
	fs := mockFilesystem{
		app:          app,
		installation: installation,
		files: map[string][]byte{
			"default/config.yaml": []byte(globalConfig),
			"default/apps/" + app + "/configmap-values.yaml.template": []byte(appConfigmapTemplate),
			"installations/" + installation + "/secrets.yaml":         []byte(installationSecret),
			"default/apps/" + app + "secret-values.yaml.template":     []byte(appSecretTemplate),
		},
	}
	return &fs
}

func (fs *mockFilesystem) AddConfigPatch(patch string) {
	p := "installations/" + fs.installation + "/config.yaml.patch"
	fs.files[p] = []byte(patch)
}

func (fs *mockFilesystem) AddConfigmapTemplatePatch(patch string) {
	p := "installations/" + fs.installation + "/apps/" + fs.app + "/configmap-values.yaml.patch.template"
	fs.files[p] = []byte(patch)
}

func (fs *mockFilesystem) AddIncludeFile(filepath, contents string) {
	p := path.Join("include", filepath)
	fs.files[p] = []byte(contents)
}

func (fs *mockFilesystem) ReadFile(filepath string) ([]byte, error) {
	v, ok := fs.files[filepath]
	if !ok {
		return []byte{}, microerror.Maskf(notFoundError, "%q not found", filepath)
	}
	return v, nil
}

type mockFile struct {
	name string
}

func (ff *mockFile) Name() string {
	return path.Base(ff.name)
}

func (ff mockFile) Size() int64 {
	return 1
}

func (ff mockFile) Mode() os.FileMode {
	return os.FileMode(0666)
}

func (ff mockFile) ModTime() time.Time {
	return time.Now()
}

func (ff mockFile) IsDir() bool {
	return false
}

func (ff mockFile) Sys() interface{} {
	return nil
}

func (fs *mockFilesystem) ReadDir(_ string) ([]os.FileInfo, error) {
	out := []os.FileInfo{}
	for k := range fs.files {
		if !strings.HasPrefix(k, "include") {
			continue
		}
		out = append(out, &mockFile{k})
	}
	return out, nil
}
