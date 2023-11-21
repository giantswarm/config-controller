# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added repository name parameter to be able to point to different config repositories, defaults to `config`

### Changed

- Removed usage of `config.giantswarm.io/version` annotation from AppCatalogEntries in favor of repository ref parameter, defaults to `main`
- Changed repository clone mechanism to pull and initialize submodules as well

### Removed

- Removed the `lint` CLI command.

## [0.9.0] - 2023-11-10

### Changed

- Add a switch for PSP CR installation.

## [0.8.0] - 2023-07-31

### Added

- Add Service Monitor.

## [0.7.0] - 2023-07-04

### Added

- Add use of the runtime/default seccomp profile.

### Removed

- Remove push to `shared-app-collection` as it is deprecated.

### Changed

- Updated default `securityContext` values to comply with PSS policies.


## [0.6.0] - 2022-07-21

### Changed

- Add deprecation notice in favor of `konfigure`.
- Replace deprecated vault decrypt call with updated method from `valuemodifier` library.

## [0.5.1] - 2021-11-30

### Fixed

- Generate missing Config CRD.

## [0.5.0] - 2021-11-29

### Changed

- Drop `apiextensions` dependency.

## [0.4.0] - 2021-08-09

## [0.3.3] - 2021-08-05

### Added

- Add a flag to allow the user to be passed to `opsctl` when generating locally.

### Fixed

- Fix fetching the latest tag for a version range.

## [0.3.2] - 2021-06-03

### Changed

- Set `unique` app configuration to `true` by default.

## [0.3.1] - 2021-06-02

### Fixed

- Fix missing new `architect-orb` version.

## [0.3.0] - 2021-06-02

### Added

- Allow for raw config generation from CLI.

### Changed

- Prepare helm values to configuration management.
- Update architect-orb to v3.0.0.

## [0.2.11] - 2021-04-16

### Fixed

- Improve lookup for paths defined in templates/patches.

## [0.2.10] - 2021-04-13

### Fixed

- Use `text/template` instead of `html` to avoid escaping strings.
- Return a more descriptive error when given invalid YAML.

## [0.2.9] - 2021-04-02

### Fixed

- Bump protobuf to v1.3.2 (CVE-2021-3121)

## [0.2.8] - 2021-03-24

### Fixed

- Prevent panic when linter cross-references apps and installations.

## [0.2.7] - 2021-03-23

### Fixed

- Skip non-existent application patches when linting.

## [0.2.6] - 2021-02-22

### Fixed

- Bring back `application/v1alpha1` API extension to the registered schemas.

## [0.2.5] - 2021-02-22


### Added

- Add `skip-validation-regexp` to skip selected fields validation.

### Deleted

- Delete `App` CR controller.


## [0.2.4] - 2021-02-16

### Added

- Add configuration linter under `lint` command.
- Add logs when generating config via CLI.
- Reconcile Config CRs.

### Fixed

- Throw errors when template keys are missing.

## [0.2.3] - 2021-01-28

### Fixed

- Add missing `giantswarm.io/monitoring-*` annotations.
- Update configuration ConfigMap ans Secret only when they change.
- Retry App CR modifications on conflicts.

## [0.2.2] - 2021-01-19

### Fixed

- Add `giantswarm.io/monitoring: "true"` label to the Service to make sure the
  app is scraped by the new monitoring platform.
- Resolve catalog URL using storage URL from AppCatalog CR rather than using
  simple format string.

## [0.2.1] - 2021-01-14

### Fixed

- Remove old ConfigMap and Secret when a new config version is set.

### Fixed

- Use `config.giantswarm.io/version` Chart annotation to determine configuration version.

## [0.2.0] - 2021-01-12

### Added
- Add `values` handler, which generates App ConfigMap and Secret.
- Allow caching tags and pulled repositories.
- Handle `app-operator.giantswarm.io/pause` annotation.
- Clear `app-operator.giantswarm.io/pause` if App CR does is not annotated with config version.
- Annotate App CR with config version defined in catalog's `index.yaml`.

## [0.1.0] - 2020-11-26

### Added

- Create CLI/daemon scaffolding.
- Generate application configuration using `generate` command.

[Unreleased]: https://github.com/giantswarm/config-controller/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/giantswarm/config-controller/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/giantswarm/config-controller/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/giantswarm/config-controller/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/giantswarm/config-controller/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/giantswarm/config-controller/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/config-controller/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/config-controller/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/giantswarm/config-controller/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/giantswarm/config-controller/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/giantswarm/config-controller/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/config-controller/compare/v0.2.11...v0.3.0
[0.2.11]: https://github.com/giantswarm/config-controller/compare/v0.2.10...v0.2.11
[0.2.10]: https://github.com/giantswarm/config-controller/compare/v0.2.9...v0.2.10
[0.2.9]: https://github.com/giantswarm/config-controller/compare/v0.2.8...v0.2.9
[0.2.8]: https://github.com/giantswarm/config-controller/compare/v0.2.7...v0.2.8
[0.2.7]: https://github.com/giantswarm/config-controller/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/giantswarm/config-controller/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/giantswarm/config-controller/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/giantswarm/config-controller/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/giantswarm/config-controller/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/giantswarm/config-controller/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/config-controller/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/config-controller/releases/tag/v0.2.0
[0.1.0]: https://github.com/giantswarm/config-controller/releases/tag/v0.1.0
