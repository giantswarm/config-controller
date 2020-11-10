package generator

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

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
			path.Join(defaultDir, defaultConfigFile):                          []byte(globalConfig),
			path.Join(defaultDir, appsSubDir, app, configmapTemplateFile):     []byte(appConfigmapTemplate),
			path.Join(installationsDir, installation, installationSecretFile): []byte(installationSecret),
			path.Join(defaultDir, appsSubDir, app, secretTemplateFile):        []byte(appSecretTemplate),
		},
	}
	return &fs
}

func (fs *mockFilesystem) AddConfigPatch(patch string) {
	p := path.Join(installationsDir, fs.installation, installationConfigPatchFile)
	fs.files[p] = []byte(patch)
}

func (fs *mockFilesystem) AddConfigmapTemplatePatch(patch string) {
	p := path.Join(
		installationsDir,
		fs.installation,
		appsSubDir,
		fs.app,
		configmapTemplatePatchFile,
	)
	fs.files[p] = []byte(patch)
}

func (fs *mockFilesystem) AddIncludeFile(filepath, contents string) {
	p := path.Join(
		includeDir,
		filepath,
	)
	fs.files[p] = []byte(contents)
}

func (fs *mockFilesystem) Exists(filepath string) bool {
	_, ok := fs.files[filepath]
	return ok
}

func (fs *mockFilesystem) ReadFile(filepath string) ([]byte, error) {
	v, ok := fs.files[filepath]
	if !ok {
		return []byte{}, fmt.Errorf("file does not exist")
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
	for k, _ := range fs.files {
		if !strings.HasPrefix(k, includeDir) {
			continue
		}
		out = append(out, &mockFile{k})
	}
	return out, nil
}

func TestGenerator_GenerateConfig(t *testing.T) {
	fs := newMockFilesystem("puma", "operator", configYaml, configmapTemplate, installationSecret, secretTemplate)
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
