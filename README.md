# ICU Cloud Native Buildpack

The ICU CNB provides [International Components for Unicode](http://site.icu-project.org/home) libraries.

## Integration

The ICU CNB provides icu as a dependency. Downstream buildpacks, like [Dotnet
Core Publish CNB](https://github.com/paketo-buildpacks/dotnet-core-publish) can
require the icu dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the ICU dependency is "icu". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "icu"

  # The version of the ICU dependency is not required. In the case it
  # is not specified, the buildpack will provide the newest version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "66.*", "66.1.*", or even
  # "66.1.0".
  version = "66.1.0"

  # The ICU CNB supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the ICU
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run ICU
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the ICU dependency is
    # available on the $PATH for the running application. If you are writing an
    # application that needs to run ICU at runtime, this flag should be set to
    # true.
    launch = true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configurations

The `icu` buildpack does not support configurations using `buildpack.yml`.

