package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type RegistryConfig struct {
	Name   string  `yaml:"name"`
	URL    string  `yaml:"url"`
	Ref    *string `yaml:"ref"`
	Target *string `yaml:"target"`
}

type ReleaseConfig struct {
	Name             string            `yaml:"name"`
	Pack             string            `yaml:"pack"`
	VarFiles         []string          `yaml:"var-files"`
	Vars             map[string]string `yaml:"vars"`
	Environments     []string          `yaml:"environments"`
	EnvironmentFiles []string          `yaml:"environment-files"`
	NomadAddr        string            `yaml:"nomad-addr"`
	NomadToken       string            `yaml:"nomad-token"`
}

type Config struct {
	Registries      []RegistryConfig         `yaml:"registries"`
	Environments    map[string]ReleaseConfig `yaml:"environments"`
	Releases        []ReleaseConfig          `yaml:"releases"`
	Path            string                   `yaml:"-"`
	NomadPackBinary string                   `yaml:"-"`
	ExtraVars       map[string]string        `yaml:"-"`
}

// WorkDir returns the directory where the packfile is located.
func (config *Config) WorkDir() string {
	path := config.Path
	return filepath.Dir(path)
}

func NewFromFile(file string, cmd *cobra.Command) (*Config, error) {
	config := Config{}

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	environment, err := cmd.Flags().GetString("environment")
	if environment != "" {
		newEnvironments := map[string]ReleaseConfig{}
		newEnvironments[environment] = config.Environments[environment]
		config.Environments = newEnvironments
	}

	release, err := cmd.Flags().GetString("release")
	if release != "" {
		newReleases := []ReleaseConfig{}
		for _, r := range config.Releases {
			if r.Name == release {
				newReleases = append(newReleases, r)
			}
		}
		config.Releases = newReleases
	}

	config.Path = file
	config.NomadPackBinary, err = cmd.Flags().GetString("nomad-pack-binary")

	// Parse --var flags
	varFlags, err := cmd.Flags().GetStringSlice("var")
	if err != nil {
		return nil, err
	}
	
	config.ExtraVars = make(map[string]string)
	for _, varFlag := range varFlags {
		// Split on first = to support values with = in them
		parts := strings.SplitN(varFlag, "=", 2)
		if len(parts) == 2 {
			config.ExtraVars[parts[0]] = parts[1]
		}
	}

	return &config, err
}
