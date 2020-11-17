package generator

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
)

func TestGenerator_GenerateConfig(t *testing.T) {
	testCases := []struct {
		name     string
		caseFile string

		app          string
		installation string
	}{
		{
			name:     "case 0 - basic templates and installation-level config.yaml patch",
			caseFile: "test/case0.yaml",

			app:          "operator",
			installation: "puma",
		},

		/*
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
		*/
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "config-controller-test")
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			defer os.RemoveAll(tmpDir)

			fs := newMockFilesystem(tmpDir, tc.caseFile)
			config := Config{Fs: fs}

			g, err := New(&config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			configmap, secret, err := g.GenerateConfig(tc.installation, tc.app)
			if err != nil {
				t.Fatalf("unexpected error: %s", microerror.Pretty(err, true))
			}
			if configmap != fs.ExpectedConfigmap {
				t.Fatalf("configmap not expected, got: %s", configmap)
			}
			if secret != fs.ExpectedSecret {
				t.Fatalf("secret not expected, got: %s", secret)
			}
		})
	}
}

type mockFilesystem struct {
	tempDirPath string

	ExpectedConfigmap string
	ExpectedSecret    string
}

type testFile struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

func newMockFilesystem(temporaryDirectory, caseFile string) *mockFilesystem {
	fs := mockFilesystem{
		tempDirPath: temporaryDirectory,
	}
	for _, p := range []string{"default", "installations", "include"} {
		if err := os.Mkdir(path.Join(temporaryDirectory, p), 0777); err != nil {
			panic(err)
		}
	}

	rawData, err := ioutil.ReadFile(caseFile)
	if err != nil {
		panic(err)
	}

	splitFiles := strings.Split(string(rawData), "---")

	for _, rawYaml := range splitFiles {
		file := testFile{}
		if err := yaml.Unmarshal([]byte(rawYaml), &file); err != nil {
			panic(err)
		}

		p := path.Join(temporaryDirectory, file.Path)
		dir, filename := path.Split(p)

		switch filename {
		case "configmap-values.yaml.golden":
			fs.ExpectedConfigmap = file.Data
			continue
		case "secret-values.yaml.golden":
			fs.ExpectedSecret = file.Data
			continue
		}

		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}

		err := ioutil.WriteFile(p, []byte(file.Data), 0777)
		if err != nil {
			panic(err)
		}
	}

	return &fs
}

func (fs *mockFilesystem) ReadFile(filepath string) ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(fs.tempDirPath, filepath))
	if err != nil {
		return []byte{}, microerror.Maskf(notFoundError, "%q not found", filepath)
	}
	return data, nil
}

func (fs *mockFilesystem) ReadDir(dirpath string) ([]os.FileInfo, error) {
	p := path.Join(fs.tempDirPath, dirpath)
	return ioutil.ReadDir(p)
}
