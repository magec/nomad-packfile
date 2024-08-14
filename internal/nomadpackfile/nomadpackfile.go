package nomadpackfile

import (
	"bytes"
	"log"
	"os"
	"slices"
	"strings"
	"text/template"

	configpkg "github.com/magec/nomad-packfile/internal/config"
	"github.com/magec/nomad-packfile/internal/nomadpack"
	"github.com/pterm/pterm"
	"go.uber.org/zap"
)

// This is a simple AST for the NomadPackFile
type NomadPackFile struct {
	config     configpkg.Config
	registries map[string]RegistryNode
	releases   []ReleaseNode
	logger     *zap.Logger
}

type RegistryNode struct {
	/// Name of the Registry this is the name that will be used in releases when you need to refer to this registry
	Name string

	/// URL of the registry
	URL string

	Ref           *string
	Target        *string
	NomadPackFile *NomadPackFile
}

type ReleaseNode struct {
	Name          string
	Pack          string
	Registry      RegistryNode
	VarFiles      []string
	Vars          map[string]string
	workDir       string
	NomadPackFile *NomadPackFile
	Environments  []string
	NomadAddr     string
	NomadToken    string
}

func (registry RegistryNode) Plan() error {
	nomadPack, err := registry.NomadPackFile.NomadPack()
	if err != nil {
		log.Fatalf("Error getting initializing nomad-pack: %s", err)
	}

	return nomadPack.AddRegistry(registry.Name, registry.URL, registry.Ref, registry.Target)
}

func (release ReleaseNode) Plan() error {
	nomadPack, err := release.nomadPack()
	if err != nil {
		log.Fatalf("Error getting initializing nomad-pack: %s", err)
	}

	return nomadPack.Plan(release.workDir, true, release.Registry.Name, release.Registry.Ref, release.Pack, release.VarFiles, release.Vars)
}

func (release ReleaseNode) Run() error {
	nomadPack, err := release.nomadPack()

	if err != nil {
		log.Fatalf("Error getting initializing nomad-pack: %s", err)
	}
	return nomadPack.Run(release.workDir, true, release.Registry.Name, release.Registry.Ref, release.Pack, release.VarFiles, release.Vars)
}

func (release ReleaseNode) Render() error {
	nomadPack, err := release.nomadPack()
	if err != nil {
		log.Fatalf("Error getting initializing nomad-pack: %s", err)
	}

	return nomadPack.Render(release.workDir, true, release.Registry.Name, release.Registry.Ref, release.Pack, release.VarFiles, release.Vars)
}

func (release ReleaseNode) nomadPack() (nomadPack *nomadpack.NomadPack, err error) {
	nomadPack, err = release.NomadPackFile.NomadPack()
	nomadPack = nomadPack.NomadAddr(release.NomadAddr).NomadToken(release.NomadToken)
	return nomadPack, err
}

func New(config configpkg.Config, logger *zap.Logger) *NomadPackFile {
	return &NomadPackFile{config: config, logger: logger, registries: make(map[string]RegistryNode)}
}

func (n *NomadPackFile) NomadPack() (*nomadpack.NomadPack, error) {
	return nomadpack.New(n.config.NomadPackBinary, n.logger)
}

func (n *NomadPackFile) Plan() error {
	for _, registry := range n.registries {
		err := registry.Plan()
		if err != nil {
			return err
		}
	}
	for _, release := range n.releases {
		err := release.Plan()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NomadPackFile) Render() error {
	for _, registry := range n.registries {
		err := registry.Plan()
		if err != nil {
			return err
		}
	}
	for _, release := range n.releases {
		err := release.Render()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NomadPackFile) Run() error {
	for _, registry := range n.registries {
		err := registry.Plan()
		if err != nil {
			return err
		}
	}
	for _, release := range n.releases {
		err := release.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

type templateEnvironmentContext struct {
	Name string
}

type templateContext struct {
	Environment templateEnvironmentContext
	Env         map[string]string
}

func (n *NomadPackFile) Compile() error {
	for _, registryConfig := range n.config.Registries {
		if n.registries[registryConfig.Name].Name != "" {
			pterm.Warning.Printf("Registry %s already exists, skipping", registryConfig.Name)
			continue
		}

		n.registries[registryConfig.Name] = RegistryNode{
			Name:          registryConfig.Name,
			URL:           registryConfig.URL,
			Ref:           registryConfig.Ref,
			Target:        registryConfig.Target,
			NomadPackFile: n,
		}
	}

	for name, environmentRelease := range n.config.Environments {
		for _, release := range n.config.Releases {
			if release.Environments != nil && !slices.Contains(release.Environments, name) {
				continue
			}

			if environmentRelease.NomadAddr != "" {
				release.NomadAddr = environmentRelease.NomadAddr
			}
			if environmentRelease.NomadToken != "" {
				release.NomadToken = environmentRelease.NomadToken
			}

			context := templateContext{
				Environment: templateEnvironmentContext{
					Name: name,
				},
				Env: environmentToHash(),
			}

			splitPack := strings.Split(release.Pack, "/")
			if len(splitPack) != 2 {
				log.Fatalf("Invalid pack name: %s", release.Pack)
			}

			if n.registries[splitPack[0]].Name == "" {
				log.Fatalf("Registry %s not found", splitPack[0])
			}

			registry := n.registries[splitPack[0]]
			pack := splitPack[1]
			workDir := n.config.WorkDir()

			var err error
			release.NomadAddr, err = execTemplate(release.NomadAddr, context)
			if err != nil {
				log.Fatalf("Error interpreting template in nomad-addr: %s, err: %s.", release.NomadAddr, err)
				panic(err)

			}
			release.NomadToken, err = execTemplate(release.NomadToken, context)
			if err != nil {
				log.Fatalf("Error interpreting template in nomad-addr: %s, err: %s.", release.NomadAddr, err)
				panic(err)
			}

			newVarFiles := []string{}
			for _, varFile := range release.VarFiles {
				newVarFile, err := execTemplate(varFile, context)
				if err != nil {
					panic(err)
				}
				newVarFiles = append(newVarFiles, newVarFile)
			}

			newVars := map[string]string{}
			for key, bar := range release.Vars {
				newVar, err := execTemplate(bar, context)
				if err != nil {
					panic(err)
				}
				newVars[key] = newVar
			}

			releaseNode := ReleaseNode{
				Name:          release.Name,
				Pack:          pack,
				Registry:      registry,
				VarFiles:      newVarFiles,
				workDir:       workDir,
				NomadPackFile: n,
				NomadAddr:     release.NomadAddr,
				NomadToken:    release.NomadToken,
				Vars:          newVars,
			}

			n.releases = append(n.releases, releaseNode)
		}
	}

	n.logger.Debug("Compiled NomadPackFile", zap.Any("registries", n.registries), zap.Any("releases", n.releases))
	return nil
}

func environmentToHash() (result map[string]string) {
	result = make(map[string]string, len(os.Environ()))
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		result[pair[0]] = pair[1]
	}
	return
}

func execTemplate(tmpl string, context templateContext) (string, error) {
	t, err := template.New("nomad-pack-template").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, context)
	if err != nil {
		return "", err
	}

	return doc.String(), nil
}
