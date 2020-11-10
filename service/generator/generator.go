package generator

import (
	"bytes"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	pathmodifier "github.com/giantswarm/valuemodifier/path"
)

const (
	/*
		- .yaml is a values source
		- .yaml.patch overrides values source
		- .yaml.template is a template
		- .yaml.template.patch overrides template

		Folder structure:
			default/
			    config.yaml
			    apps/
			        aws-operator/
				        ...
				    azure-operator/
				        configmap-values.yaml.template
				        secret-values.yaml.template
			installations/
			    ghost/
				    ...
				godsmack/
				    config.yaml.patch
					secrets.yaml
					apps/
					    azure-operator/
						    configmap-values.yaml.template.patch
	*/

	// global directories (at /)
	defaultDir       = "default"
	includeDir       = "include"
	installationsDir = "installations"

	// both in /defaultDir and /installationsDir/<installation>
	appsSubDir = "apps"

	// installation-level config
	defaultConfigFile           = "config.yaml"
	installationConfigPatchFile = defaultConfigFile + ".patch"
	installationSecretFile      = "secrets.yaml"

	// default app-level config
	configmapTemplateFile = "configmap-values.yaml.template"
	secretTemplateFile    = "secret-values.yaml.template"

	// installation app-level config
	configmapTemplatePatchFile = "configmap-values.yaml.template.patch"
)

var (
	funcMap = sprig.FuncMap()
)

type Filesystem interface {
	Exists(string) bool
	ReadFile(string) ([]byte, error)
	ReadDir(string) ([]os.FileInfo, error)
}

type Config struct {
	App          string
	Fs           Filesystem
	Installation string
	Version      string
}

type Generator struct {
	app          string
	fs           Filesystem
	installation string
	version      string
}

func New(config *Config) (*Generator, error) {
	g := Generator{
		app:          config.App,
		fs:           config.Fs,
		installation: config.Installation,
		version:      config.Version,
	}
	return &g, nil
}

// GenerateConfig creates final configmap values and secret values for helm to
// use by performing the following operations:
// 1. Get configmap context and patch it with installation-specific overrides (if
//    available)
// 2. Get app-specific configmap template and patch it with installation-specific
//    app overrides (if available)
// 3. Render app configmap template (result of 2.) with context values (result
//    of 1.)
// 4. Get secrets context
// 5. Get app-specific secrets template.
// 6. Render app secrets template (result of 4.) with installation secrets (result of 5.)
func (g Generator) GenerateConfig() (string, string, error) {
	// 1.
	configmapContext, err := g.getWithPatchIfExists(
		path.Join(defaultDir, defaultConfigFile),
		path.Join(installationsDir, g.installation, installationConfigPatchFile),
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 2.
	configmapTemplate, err := g.getWithPatchIfExists(
		path.Join(defaultDir, appsSubDir, g.app, configmapTemplateFile),
		path.Join(installationsDir, g.installation, appsSubDir, g.app, configmapTemplatePatchFile),
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 3.
	configmap, err := g.renderTemplate(configmapTemplate, configmapContext)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 4.
	secretsContext, err := g.getWithPatchIfExists(
		path.Join(installationsDir, installationSecretFile),
		"",
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 5.
	secretsTemplate, err := g.getWithPatchIfExists(
		path.Join(defaultDir, appsSubDir, g.app, secretTemplateFile),
		"",
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 6.
	secrets, err := g.renderTemplate(secretsTemplate, secretsContext)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	return configmap, secrets, nil
}

// getWithPatchIfExists provides contents of filepath overwritten by patch at
// patchFilepath. File at patchFilepath may be non-existent, resulting in pure
// file at filepath being returned.
func (g Generator) getWithPatchIfExists(filepath, patchFilepath string) (string, error) {
	var err error

	var base []byte
	{
		base, err = g.fs.ReadFile(filepath)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	// patch is not obligatory
	var patch []byte
	if patchFilepath != "" && g.fs.Exists(patchFilepath) {
		patch, err = g.fs.ReadFile(patchFilepath)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	result, err := applyPatch(base, patch)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return result, nil
}

func applyPatch(base, patch []byte) (string, error) {
	var basePathSvc *pathmodifier.Service
	{
		c := pathmodifier.DefaultConfig()
		c.InputBytes = base
		svc, err := pathmodifier.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
		basePathSvc = svc
	}

	var patchPathSvc *pathmodifier.Service
	{
		c := pathmodifier.DefaultConfig()
		c.InputBytes = patch
		svc, err := pathmodifier.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
		patchPathSvc = svc
	}

	patchedPaths, err := patchPathSvc.All()
	if err != nil {
		return "", microerror.Mask(err)
	}

	for _, p := range patchedPaths {
		value, err := patchPathSvc.Get(p)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = basePathSvc.Set(p, value)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	outputBytes, err := basePathSvc.OutputBytes()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(outputBytes), nil
}

func (g Generator) renderTemplate(templateText string, context string) (string, error) {
	c := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(context), &c)
	if err != nil {
		return "", microerror.Mask(err)
	}

	t := template.Must(template.New("main").Funcs(funcMap).Parse(templateText))

	// Add include files as templates usable by calling
	// `{{ template "name" . }}` in templateText.
	files, err := g.fs.ReadDir(includeDir)
	if err != nil {
		return "", microerror.Mask(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		baseName := path.Base(file.Name())
		if strings.ContainsRune(baseName, '.') {
			baseName = strings.SplitN(baseName, ".", 1)[0]
		}

		contents, err := g.fs.ReadFile(
			path.Join(includeDir, file.Name()),
		)
		if err != nil {
			return "", microerror.Mask(err)
		}

		_, err = t.New(baseName).Funcs(funcMap).Parse(string(contents))
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	// render final template
	out := bytes.NewBuffer([]byte{})
	err = t.Execute(out, c)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return out.String(), nil
}
