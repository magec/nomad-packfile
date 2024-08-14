# Nomad-Packfile
Declaratively deploy nomad packs to different clusters.

## Introduction

[Nomad Pack](https://github.com/hashicorp/nomad-pack) is a a templating and
packaging tool used with [HashiCorp Nomad](https://www.nomadproject.io/).

It is used to deploy applications to Nomad using templates, this way
packages can be reused and parametrized.

## Why another tool on top of that?
Even though packages are declarative by themselves, their *instantaions* are not.

This is because `nomad-pack` permits you to set up variables when running, etc. Moreover,
you cannot define a set of packages to be deployed to the same or different environments.

This is where `nomad-packfile` comes into place, it can be used to:

- Declare different environments with different global configurations.
- Declare several releases with their values set (optionally you can still use ENV_VARS).

It is strongly inspired by [Helmfile](https://github.com/helmfile/helmfile).

## Installation

Download the binary from [release](https://github.com/magec/nomad-packfile/releases) and 
put it in your PATH.

## Configuration

The tool reads a `packfile.yaml` file in the current directory (or the one specified with `-f`),
and synchronize the desired state with the Nomad cluster(s).

Here is an example of a `packfile.yaml`:

```yaml
---
environments:
  staging:
    nomad-addr: https://staging.nomad.cluster
    nomad-token: "{{ .Env.STAGING_NOMAD_TOKEN }}"
  production:
    nomad-addr: https://production.nomad.cluster
    nomad-token: "{{ .Env.PRODUCTION_NOMAD_TOKEN }}"

registries:
  - name: myorg
    url: github.com/myorg/nomad-packs

releases:
  - name: application
    pack: myorg/some_application
    vars:
      image_tag: "{{ .Env.IMAGE_TAG }}"
    var-files:
      - nomad/common.hcl
      - "nomad/{{ .Environment.Name }}.hcl"
```

In this example, we have two environments, `staging` and `production`, and a release `application`. Nomad
pack will be deployed to both clusters with the configuration specified in the `vars` and `var-files` sections.
Note that you can use templates in the `vars` and `var-files` sections.

### Environments

The `environments` section is used to define the different environments where the packs will be deployed. It
can be any number of them. Once you have an environment defined, everytime you execute `nomad-packfile` the desired state
will be the the product of releases and environments. Note that the resulting configuration for each release, will be the
environment configuration merged with the release one.

This means that for the example, you end up with these two releases to be deployed:

```yaml
...
  - name: application
    pack: myorg/some_application
    nomad-addr: https://staging.nomad.cluster        # Inherited from Environment config
    nomad-token: "{{ .Env.STAGING_NOMAD_TOKEN }}"    # Inherited from Environment config
    vars:
      image_tag: "{{ .Env.IMAGE_TAG }}"
    var-files:
      - nomad/common.hcl
      - nomad/staging.hcl                            # Template resolved to environment name.
  - name: application
    pack: myorg/some_application
    nomad-addr: https://production.nomad.cluster     # Inherited from Environment config
    nomad-token: "{{ .Env.PRODUCTION_NOMAD_TOKEN }}" # Inherited from Environment config
    vars:
      image_tag: "{{ .Env.IMAGE_TAG }}"
    var-files:
      - nomad/common.hcl
      - nomad/production.hcl"                        # Template resolved to environment name.
...
```

This means that in this example, you will deploy the application to
both `staging` and `production` environments. You can filter out the environments by using the `--environment` flag.

### Registries
The `registries` section is used to define the registries where the packs are stored. You can define as many registries as you want.
They will be referenced in releases when you declare the name of the pack to see (see bellow).

Registries can have:

- **name**: The name of the registry, it will be used in releases to reference this registry. It will also be used when fetching the registries.
- **url**: Url of the registry (see `nomad-pack registry add --help` for more information).
- **Ref**: Specific git ref of the registry or pack to be added. Supports tags,
        SHA, and latest. If no ref is specified, defaults to latest. Running
        "nomad registry add" multiple times for the same ref is idempotent,
        however running "nomad-pack registry add" without specifying a
        ref, or when specifying @latest, is destructive, and will overwrite
        current @latest in the global cache. Using ref with a file path is not
        supported.
- **Target**: A specific pack within the registry to be added.

### Releases
This is where you define the releases themselves. It will reference the pack and also permits declaring variable files and variables. Note
that you can use templates inside these values.

Releases can have:

- **name**: The name of the release.
- **pack**: The reference of the pack in the form of. By default it will treat it as a path, if you want to reference a registry, you need to use
            `registry://registry_name/pack`.
- **var-files**: An array of varfiles to be added to command invocation. If files are not found it will show a warning and skip it.
- **vars**: An array of vars to be added to `nomad-pack` command invocation.
- **environments**: This permits filtering out environments in case you don't want a given release to be deployed to every environment.
- **nomad-addr**: Nomad addr to be used to deploy. This is usually set in the environment configuration.
- **nomad-token**: Nomad Token to be used to deploy. This is usually set in the environment configuration using an templating and an env var.

#### Templating
As mentioned, you can use templating in (`nomad-addr`, `nomad-token`, `var-files` and `vars`). This way, you can customize the configuration
based environment variables or the name of the environment (as shown in the example you can reference the Environment using `Environment.Name`).
This allows a more clean setup and less repetition.

## Usage
```bash
Declare the desired state of your packs and let nomad-packfile synchronize it with your Nomad cluster.

Usage:
  nomad-packfile [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  plan        Execute a nomad-plan for every pack in the desired state
  render      Execute a nomad-render for every pack in the desired state
  run         Execute a nomad-run for every pack in the desired state

Flags:
      --environment string         Specify the environment name.
  -f, --file string                Load config from file or directory (default "packfile.yaml")
  -h, --help                       help for nomad-packfile
      --log-level string           Log Level. (default "fatal")
      --nomad-pack-binary string   Path to the nomad-pack binary. (default "nomad-pack")
      --release string             Specify the release (this filters out any release apart from the specified one).

Use "nomad-packfile [command] --help" for more information about a command.
```

`nomad-packfile` currently allows three commands:

- **plan**: This will execute a `nomad-pack plan` for every release in the desired state.
- **render**: This will execute a `nomad-pack render` for every release in the desired state.
- **run**: This will execute a `nomad-pack run` for every release in the desired state.
