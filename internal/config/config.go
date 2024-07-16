package config

import (
	"os"
	"path/filepath"

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
	Name         string            `yaml:"name"`
	Pack         string            `yaml:"pack"`
	VarFiles     []string          `yaml:"var-files"`
	Vars         map[string]string `yaml:"vars"`
	Environments []string          `yaml:"environments"`
	NomadAddr    string            `yaml:"nomad-addr"`
	NomadToken   string            `yaml:"nomad-token"`
}

type Config struct {
	Registries      []RegistryConfig         `yaml:"registries"`
	Environments    map[string]ReleaseConfig `yaml:"environments"`
	Releases        []ReleaseConfig          `yaml:"releases"`
	Path            string                   `yaml:"-"`
	NomadPackBinary string                   `yaml:"-"`
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
	config.Path = file
	config.NomadPackBinary, err = cmd.Flags().GetString("nomad-pack-binary")

	return &config, err
}
