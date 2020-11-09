package generator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	pathmodifier "github.com/giantswarm/valuemodifier/path"
)

// TODO:
// - Dependency tree
// - Given a dir and app, version, installation, find all the files
// - Be really vebose on how the config is merged
// - Write a nice struct which basically solves the merge tree.

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

type Config struct {
	App          string
	Dir          string
	Installation string
	Version      string
}

type Generator struct {
	app          string
	dir          string
	installation string
	version      string
}

func New(config *Config) (*Generator, error) {
	g := Generator{
		app:          config.App,
		dir:          config.Dir,
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
	configmapContext, err := getWithPatchIfExists(
		path.Join(g.dir, defaultDir, defaultConfigFile),
		path.Join(g.dir, installationsDir, g.installation, installationConfigPatchFile),
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 2.
	configmapTemplate, err := getWithPatchIfExists(
		path.Join(g.dir, defaultDir, appsSubDir, g.app, configmapTemplateFile),
		path.Join(g.dir, installationsDir, g.installation, appsSubDir, g.app, configmapTemplatePatchFile),
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
	secretsContext, err := getWithPatchIfExists(
		path.Join(g.dir, installationsDir, installationSecretFile),
		"",
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 5.
	secretsTemplate, err := getWithPatchIfExists(
		path.Join(g.dir, defaultDir, appsSubDir, g.app, secretTemplateFile),
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
func getWithPatchIfExists(filepath, patchFilepath string) (string, error) {
	var base string
	{
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			return "", microerror.Mask(err)
		}
		base = string(data)
	}

	var patch string
	if patchFilepath != "" {
		data, err := ioutil.ReadFile(filepath)
		// patch is not obligatory
		if err != nil && !os.IsNotExist(err) {
			return "", microerror.Mask(err)
		}
		patch = string(data)
	}

	result, err := applyPatch(base, patch)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return result, nil
}

func applyPatch(base, patch string) (string, error) {
	var basePathSvc *pathmodifier.Service
	{
		c := pathmodifier.DefaultConfig()
		c.InputBytes = []byte(base)
		svc, err := pathmodifier.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
		basePathSvc = svc
	}

	var patchPathSvc *pathmodifier.Service
	{
		c := pathmodifier.DefaultConfig()
		c.InputBytes = []byte(patch)
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

	out := bytes.NewBuffer([]byte{})
	fMap := template.FuncMap{"include": g.include}
	t := template.Must(template.New("values").Funcs(fMap).Parse(templateText))
	err = t.Execute(out, c)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return out.String(), nil
}

func (g Generator) include(filename string, indentSpaces int) (string, error) {
	filepath := path.Join(g.dir, includeDir, filename)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", microerror.Mask(err)
	}

	obj := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	data, err = yaml.Marshal(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return indent(string(data), indentSpaces), nil
}

func indent(text string, spaces int) string {
	lines := strings.Split(text, "\n")
	prefix := strings.Repeat(" ", spaces)
	out := []string{}
	for _, l := range lines {
		out = append(out, prefix+l)
	}
	return strings.Join(out, "\n")
}
