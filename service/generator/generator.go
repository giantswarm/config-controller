package generator

import (
	"bytes"
	"fmt"
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
	configmapTemplatePatchFile = "configmap-values.yaml.patch.template"
)

type Filesystem interface {
	Exists(string) bool
	ReadFile(string) ([]byte, error)
	ReadDir(string) ([]os.FileInfo, error)
}

type Config struct {
	Fs Filesystem
}

type Generator struct {
	fs Filesystem
}

func New(config *Config) (*Generator, error) {
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	g := Generator{
		fs: config.Fs,
	}

	return &g, nil
}

// GenerateConfig creates final configmap values and secret values for helm to
// use by performing the following operations:
// 1. Get configmap context and patch it with installation-specific overrides (if
//    available)
// 2. Get app-specific configmap template and render it with context (result of 1.)
// 3. Get installation-specific configmap template patch and render it with
//    context (result of 1.)
// 4. Patch app-specific template (result of 2.) with and patch it with
//    installation-specific (result of 3.) app overrides (if available)
// 5. Get secrets context
// 6. Get app-specific secrets template.
// 7. Render app secrets template (result of 4.) with installation secrets
//    (result of 5.)
func (g Generator) GenerateConfig(installation, app string) (string, string, error) {
	// 1.
	configmapContext, err := g.getWithPatchIfExists(
		path.Join(defaultDir, defaultConfigFile),
		path.Join(installationsDir, installation, installationConfigPatchFile),
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 2.
	configmapBase, err := g.getRenderedTemplate(
		path.Join(defaultDir, appsSubDir, app, configmapTemplateFile),
		configmapContext,
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 3.
	var configmapPatch string
	{
		filepath := path.Join(installationsDir, installation, appsSubDir, app, configmapTemplatePatchFile)
		if g.fs.Exists(filepath) {
			patch, err := g.getRenderedTemplate(filepath, configmapContext)
			if err != nil {
				return "", "", microerror.Mask(err)
			}
			configmapPatch = patch
		} else {
			configmapPatch = ""
		}
	}

	// 4.
	configmap, err := applyPatch(
		[]byte(configmapBase),
		[]byte(configmapPatch),
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 5.
	secretsContext, err := g.getWithPatchIfExists(
		path.Join(installationsDir, installation, installationSecretFile),
		"",
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 6.
	secretsTemplate, err := g.getWithPatchIfExists(
		path.Join(defaultDir, appsSubDir, app, secretTemplateFile),
		"",
	)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// 7.
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
	if patchFilepath == "" || !g.fs.Exists(patchFilepath) {
		return string(base), nil
	}

	var patch []byte
	{
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

func (g Generator) getRenderedTemplate(filepath, context string) (string, error) {
	templateBytes, err := g.fs.ReadFile(filepath)
	if err != nil {
		return "", microerror.Mask(err)
	}

	result, err := g.renderTemplate(string(templateBytes), context)
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

	funcMap := sprig.FuncMap()
	funcMap["include"] = g.include

	t := template.New("main").Funcs(funcMap)
	err = g.addIncludeFilesToTemplate(t)
	if err != nil {
		return "", microerror.Mask(err)
	}
	t = template.Must(t.Parse(templateText))

	// render final template
	out := bytes.NewBuffer([]byte{})
	err = t.Execute(out, c)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return out.String(), nil
}

func (g Generator) addIncludeFilesToTemplate(t *template.Template) error {
	files, err := g.fs.ReadDir(includeDir)
	if err != nil {
		return microerror.Mask(err)
	}

	funcMap := sprig.FuncMap()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		baseName := path.Base(file.Name())
		if strings.ContainsRune(baseName, '.') {
			elements := strings.SplitN(baseName, ".", 2)
			baseName = elements[0]
		}

		contents, err := g.fs.ReadFile(
			path.Join(includeDir, file.Name()),
		)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = t.New(baseName).Funcs(funcMap).Parse(string(contents))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (g Generator) include(templateName string, context interface{}) (string, error) {
	t := template.New("render-" + templateName).Funcs(sprig.FuncMap())
	err := g.addIncludeFilesToTemplate(t)
	if err != nil {
		return "", microerror.Mask(err)
	}

	templateText := fmt.Sprintf("{{ template %q . }}", templateName)

	t = template.Must(t.Parse(templateText))

	out := bytes.NewBuffer([]byte{})
	err = t.Execute(out, context)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return out.String(), nil
}
