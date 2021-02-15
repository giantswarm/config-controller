package lint

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/generator"
)

const (
	overshadowErrorThreshold float64 = 0.75
)

type LinterFunc func(d *discovery) (messages LinterMessages)

var AllLinterFunctions = []LinterFunc{
	LintUnusedconfigValues,
	LintDuplicateconfigValues,
	LintovershadowedconfigValues,
	LintUnusedConfigPatchValues,
	LintUndefinedTemplateValues,
	LintUndefinedTemplatePatchValues,
	LintUnusedSecretValues,
	LintUndefinedSecretTemplateValues,
	LintUndefinedSecretTemplatePatchValues,
	LintUnencryptedSecretValues,
	LintIncludeFiles,
}

type Config struct {
	Store           generator.Filesystem
	FilterFunctions []string
	OnlyErrors      bool
	MaxMessages     int
}

type Linter struct {
	discovery   *discovery
	funcs       []LinterFunc
	onlyErrors  bool
	maxMessages int
}

func New(c Config) (*Linter, error) {
	discovery, err := newDiscovery(c.Store)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	l := &Linter{
		discovery:   discovery,
		funcs:       GetFilteredLinterFunctions(c.FilterFunctions),
		onlyErrors:  c.OnlyErrors,
		maxMessages: c.MaxMessages,
	}

	return l, nil
}

func (l *Linter) Lint(ctx context.Context) (messages LinterMessages) {
	fmt.Printf("Linting using %d functions\n\n", len(l.funcs))
	for _, f := range l.funcs {
		singleFuncMessages := f(l.discovery)
		sort.Sort(singleFuncMessages)

		for _, msg := range singleFuncMessages {
			if l.onlyErrors && !msg.IsError() {
				continue
			}
			messages = append(messages, msg)

			if l.maxMessages > 0 && len(messages) >= l.maxMessages {
				return messages
			}
		}
	}
	return messages
}

func LintDuplicateconfigValues(d *discovery) (messages LinterMessages) {
	for path, defaultPath := range d.Config.paths {
		for _, overshadowingPatch := range defaultPath.overshadowedBy {
			patchedPath := overshadowingPatch.paths[path]
			if reflect.DeepEqual(defaultPath.value, patchedPath.value) {
				messages = append(
					messages,
					NewError(overshadowingPatch.filepath, path, "is duplicate of the same path in %s", d.Config.filepath),
				)
			}
		}
	}
	return messages
}

func LintovershadowedconfigValues(d *discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 {
		return // avoid division by 0
	}
	for path, configValue := range d.Config.paths {
		if len(configValue.overshadowedBy) == len(d.Installations) {
			messages = append(
				messages,
				NewError(d.Config.filepath, path, "is overshadowed by all config.yaml.patch files"),
			)
		} else if float64(len(configValue.overshadowedBy)/len(d.Installations)) >= overshadowErrorThreshold {
			msg := NewMessage(
				d.Config.filepath, path, "is overshadowed by %d/%d patches",
				len(configValue.overshadowedBy), len(d.Installations),
			).WithDescription("consider removing it from %s", d.Config.filepath)
			messages = append(messages, msg)
		}
	}
	return messages
}

func LintUnusedConfigPatchValues(d *discovery) (messages LinterMessages) {
	for _, configPatch := range d.ConfigPatches {
		if len(d.AppsPerInstallation[configPatch.installation]) == 0 {
			continue // avoid division by 0
		}
		for path, configValue := range configPatch.paths {
			if len(configValue.usedBy) > 0 {
				continue
			}
			messages = append(messages, NewError(configPatch.filepath, path, "is unused"))
		}
	}
	return messages
}

func LintUnusedconfigValues(d *discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 || len(d.Apps) == 0 {
		return // what's the point, nothing is defined
	}
	for path, configValue := range d.Config.paths {
		if len(configValue.usedBy) == 0 {
			messages = append(messages, NewError(d.Config.filepath, path, "is unused"))
		} else if len(configValue.usedBy) == 1 {
			msg := NewMessage(d.Config.filepath, path, "is used by just one app: %s", configValue.usedBy[0].app).
				WithDescription("consider moving this value to %s template or template patch", configValue.usedBy[0].app)
			messages = append(messages, msg)
		}
	}
	return messages
}

func LintUnusedSecretValues(d *discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 {
		return // what's the point, nothing is defined
	}
	for _, secretFile := range d.Secrets {
		for path, configValue := range secretFile.paths {
			if len(configValue.usedBy) == 0 {
				messages = append(messages, NewError(secretFile.filepath, path, "is unused"))
			} else if len(configValue.usedBy) == 1 {
				msg := NewMessage(secretFile.filepath, path, "is used by just one app: %s", configValue.usedBy[0].app).
					WithDescription("consider moving this value to %s secret-values patch", configValue.usedBy[0].app)
				messages = append(messages, msg)
			}

		}
	}
	return messages
}

func LintUndefinedSecretTemplateValues(d *discovery) (messages LinterMessages) {
	for _, template := range d.SecretTemplates {
		for path, value := range template.values {
			if !value.mayBeMissing {
				continue
			}

			messages = append(messages, NewError(template.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintUndefinedSecretTemplatePatchValues(d *discovery) (messages LinterMessages) {
	for _, template := range d.SecretTemplatePatches {
		for path, value := range template.values {
			if !value.mayBeMissing {
				continue
			}

			messages = append(messages, NewError(template.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintUndefinedTemplateValues(d *discovery) (messages LinterMessages) {
	for _, template := range d.Templates {
		for path, value := range template.values {
			if !value.mayBeMissing {
				continue
			}

			used := false
			for _, templatePatch := range d.TemplatePatches {
				if _, ok := templatePatch.paths[path]; ok {
					used = true
					break
				}

				for templatePatchPath := range templatePatch.paths {
					if strings.HasPrefix(templatePatchPath, path+".") {
						used = true
						break
					}
				}

				if used {
					break
				}
			}

			if used {
				continue
			}
			messages = append(messages, NewError(template.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintUnencryptedSecretValues(d *discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 {
		return // what's the point, nothing is defined
	}
	for _, secretFile := range d.Secrets {
		for path, configValue := range secretFile.paths {
			stringValue, ok := (configValue.value).(string)
			if !ok {
				continue
			}
			if !strings.HasPrefix(stringValue, "vault:v1:") {
				messages = append(
					messages,
					NewError(secretFile.filepath, path, "is not encrypted with Vault").
						WithDescription("valid secret values are encrypted with installation Vault's token and start with \"vault:v1:\" prefix"),
				)
			}
		}
	}
	return messages
}

func LintUndefinedTemplatePatchValues(d *discovery) (messages LinterMessages) {
	for _, templatePatch := range d.TemplatePatches {
		for path, value := range templatePatch.values {
			if !value.mayBeMissing {
				continue
			}
			messages = append(messages, NewError(templatePatch.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintIncludeFiles(d *discovery) (messages LinterMessages) {
	used := map[string]bool{}
	exist := map[string]bool{}
	for _, includeFile := range d.Include {
		exist[includeFile.filepath] = true
		for _, filepath := range includeFile.includes {
			used[filepath] = true
		}
	}

	for _, template := range d.Templates {
		for _, filepath := range template.includes {
			used[filepath] = true
		}
	}

	for _, templatePatch := range d.TemplatePatches {
		for _, filepath := range templatePatch.includes {
			used[filepath] = true
		}
	}

	if reflect.DeepEqual(exist, used) {
		return messages
	}

	for filepath := range exist {
		if _, ok := used[filepath]; !ok {
			messages = append(messages, NewError(filepath, "*", "is never included"))
		}
	}

	for filepath := range used {
		if _, ok := exist[filepath]; !ok {
			messages = append(messages, NewError(filepath, "*", "is included but does not exist"))
		}
	}

	return messages
}

//------ helper funcs -------
func GetFilteredLinterFunctions(filters []string) []LinterFunc {
	if len(filters) == 0 {
		return AllLinterFunctions
	}

	functions := []LinterFunc{}
	for _, function := range AllLinterFunctions {
		name := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
		for _, filter := range filters {
			re := regexp.MustCompile(filter)
			if re.MatchString(name) {
				functions = append(functions, function)
				break
			}
		}
	}

	return functions
}
