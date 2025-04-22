# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2025-04-22

### Changed

* Fix panic when parsing multi-doc YAML, where first doc is empty (#51)
* Bump Go to 1.24
* Bump all dependencies to the latest versions

## [0.4.0] - 2025-01-21

### Added

* `--indent` option

### Changed

* Bump all dependencies to the latest versions (Kubernetes v0.32.1)

## [0.3.0] - 2022-06-07
### Changed
* Bump to Kubernetes v0.27.0 and other dependencies
* Add simple test to verify basis functionality

## [0.2.1] - 2022-09-08
### Changed
* Bump dependencies, including Kubernetes v0.25.0

## [0.2.0] - 2022-04-14
### Changed
* Require subcommands like `create`, in order to allow others in the future

## [0.1.0] - 2021-12-25
### Added
* First release!
