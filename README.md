# Nomad-Packfile
Declaratively deploy nomad packs to different clusters.

## Introduction

[Nomad Pack](https://github.com/hashicorp/nomad-pack) is a a templating and
packaging tool used with [HashiCorp Nomad](https://www.nomadproject.io/).

It is used to deploy applications to Nomad using templates, this way
packages can be reused and parametrized.

## Why another tool on top of that?
Even though packages are declarative by themselves, their 'inistantaions' are not.
This is because `nomad-pack` permits you to set up variables when running, etc. Moreover,
You cannot define a set of packages of deploy to different environments.

This is where `nomad-packfile` comes into place, it can be used to:

- Declare different environments with different global configurations.
- Declare several releases with their values set.

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
      - "nomad{{ .Environment.Name }}.hcl"
```

In this example, we have two environments, `staging` and `production`, and a release `application`. Nomad
pack will be deployed to both clusters with the configuration specified in the `vars` and `var-files` sections.
Note that you can use templates in the `vars` and `var-files` sections.

### Environments
The `environments` section is used to define the different environments where the packs will be deployed. It
can be any number of them. Once you have an environment defined, everytime you execute `nomad-packfile` the desired state
will be the the product of releases and environments.

This means that in this example, you will deploy the application to
both `staging` and `production` environments. You can filter out the environments by using the `--environment` flag.

Also, you can limit a given release to a specific environment by setting its `environments` field to the names of the environments.

### Registries
The `registries` section is used to define the registries where the packs are stored. You can define as many registries as you want.
They will be referenced in 


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






