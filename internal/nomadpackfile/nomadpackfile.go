package nomadpackfile

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
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

type Pack struct {
	Name     string
	Registry *RegistryNode
}

func (p Pack) NomadPackCmdOpts() (params []string) {
	if p.Registry != nil {
		params = append(params, "--registry")
		params = append(params, p.Registry.Name)
		if p.Registry.Ref != nil {
			params = append(params, "--ref")
			params = append(params, *p.Registry.Ref)
		}
	}
	params = append(params, p.Name)

	return params
}

type ReleaseNode struct {
	Name          string
	Pack          Pack
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

	return nomadPack.Plan(release.workDir, true, release.VarFiles, release.Vars, release.Pack.NomadPackCmdOpts())
}

func (release ReleaseNode) Run() error {
	nomadPack, err := release.nomadPack()

	if err != nil {
		log.Fatalf("Error getting initializing nomad-pack: %s", err)
	}
	return nomadPack.Run(release.workDir, true, release.VarFiles, release.Vars, release.Pack.NomadPackCmdOpts())
}

func (release ReleaseNode) Render() error {
	nomadPack, err := release.nomadPack()
	if err != nil {
		log.Fatalf("Error getting initializing nomad-pack: %s", err)
	}

	return nomadPack.Render(release.workDir, true, release.VarFiles, release.Vars, release.Pack.NomadPackCmdOpts())
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
			workDir := n.config.WorkDir()
			fmt.Println("Release: ", release.Name)
			fmt.Println("Environment: ", name)
			fmt.Println("Release.Envirnoments: ", release.Environments)
			fmt.Println("workDir: ", workDir)

			if release.Environments != nil && !slices.Contains(release.Environments, name) {
				continue
			}

			if environmentRelease.NomadAddr != "" {
				release.NomadAddr = environmentRelease.NomadAddr
			}
			if environmentRelease.NomadToken != "" {
				release.NomadToken = environmentRelease.NomadToken
			}
			if release.EnvironmentFiles != nil {
				for _, envFile := range release.EnvironmentFiles {
					filePath := workDir + "/" + envFile
					err := godotenv.Load(filePath)
					if err != nil {
						log.Fatalf("Error reading environment file %s: %s", envFile, err)
					}
				}
			}

			context := templateContext{
				Environment: templateEnvironmentContext{
					Name: name,
				},
				Env: environmentToHash(),
			}
			var pack Pack

			if strings.HasPrefix(release.Pack, "registry://") {
				release.Pack = strings.TrimPrefix(release.Pack, "registry://")
				splitPack := strings.Split(release.Pack, "/")
				if len(splitPack) != 2 {
					log.Fatalf("Invalid pack name: %s", release.Pack)
				}

				if n.registries[splitPack[0]].Name == "" {
					log.Fatalf("Registry %s not found", splitPack[0])
				}

				registry := n.registries[splitPack[0]]
				pack = Pack{
					Registry: &registry,
					Name:     splitPack[1],
				}
			} else {
				pack = Pack{
					Name: release.Pack,
				}
			}

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
				filePath := workDir + "/" + newVarFile
				if _, err := os.Stat(filePath); err == nil {
					newVarFiles = append(newVarFiles, newVarFile)
				} else {
					pterm.Warning.Printf("Var file %s not found, skipping", filePath)
				}
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
