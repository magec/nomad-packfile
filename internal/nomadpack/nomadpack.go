package nomadpack

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/pterm/pterm"
	"go.uber.org/zap"
)

// NomadPack is a wrapper around the Nomad binary that provides a way to interact with Nomad using the Nomad Pack CLI.
type NomadPack struct {
	binaryPath string
	logger     *zap.Logger
	nomadAddr  string
	nomadToken string
}

// Creates a new NomadPack instance by providing the path to the Nomad binary.
// If the binary is not found or is not executable an error is returned.
func New(binaryPath string, logger *zap.Logger) (*NomadPack, error) {
	binaryPath, err := exec.LookPath(binaryPath)
	if err != nil {
		return nil, err
	}

	return &NomadPack{binaryPath: binaryPath, logger: logger}, nil
}

func (nomadPack *NomadPack) NomadAddr(nomad_addr string) *NomadPack {
	nomadPack.nomadAddr = nomad_addr
	return nomadPack
}

func (nomadPack *NomadPack) NomadToken(nomad_token string) *NomadPack {
	nomadPack.nomadToken = nomad_token
	return nomadPack
}

// Add nomad pack registries
// name: the name of the registry.
// source: the source of the registry.
// ref: speficic git ref of the registry or pack to be added.
// target: A specific pack within the registry to be added.
func (nomadPack *NomadPack) AddRegistry(name, source string, ref, target *string) error {
	params := []string{"registry", "add", name, source}
	if ref != nil {
		params = append(params, "--ref")
		params = append(params, *ref)
	}

	if target != nil {
		params = append(params, "--target")
		params = append(params, *target)
	}

	pterm.DefaultBasicText.Println("Adding registry", name, source)
	cmd := exec.Command(nomadPack.binaryPath, params...)

	_, err := nomadPack.runCommand(cmd)
	if err != nil {
		pterm.DefaultBasicText.Println("Successfully added.")
	}
	return err
}

// Plan runs the Nomad Pack plan command.
// workDir: the directory nomad-pack will be run in.
// diff: whether to show the diff.
// ref: speficic git ref of the registry or pack to be added.
// registry: the registry to use.
// pack: the pack to run.
// varFiles: an array of var files to use.
// vars: a map of vars to use.
func (nomadPack *NomadPack) Plan(workDir string, diff bool, varFiles []string, vars map[string]string, extraParams []string) error {
	err := nomadPack.ensureValidAuth()
	if err != nil {
		pterm.Error.Printf("Could not connect to Nomad Server: %v\n", err)
		return err
	}
	params := []string{"plan", "--diff", "--exit-code-makes-changes=0"}

	for _, varFile := range varFiles {
		params = append(params, "-var-file")
		params = append(params, varFile)
	}

	for key, value := range vars {
		params = append(params, "-var")
		params = append(params, key+"="+value)
	}
	params = append(params, extraParams...)

	cmd := exec.Command(nomadPack.binaryPath, params...)
	cmd.Dir = workDir

	pterm.DefaultBasicText.Println("Running Plan.")
	stdout, err := nomadPack.runCommand(cmd)
	if err != nil {
		pterm.DefaultBasicText.Println("Plan successfully ran.")
	}
	pterm.Println(stdout)
	return err
}

func (nomadPack *NomadPack) Run(workDir string, diff bool, varFiles []string, vars map[string]string, extraParams []string) error {
	err := nomadPack.ensureValidAuth()
	if err != nil {
		pterm.Error.Printf("Could not connect to Nomad Server: %v\n", err)
		return err
	}
	params := []string{"run"}
	for _, varFile := range varFiles {
		params = append(params, "-var-file")
		params = append(params, varFile)
	}

	for key, value := range vars {
		params = append(params, "-var")
		params = append(params, key+"="+value)
	}
	params = append(params, extraParams...)
	cmd := exec.Command(nomadPack.binaryPath, params...)
	cmd.Dir = workDir

	pterm.DefaultBasicText.Println("Running Run.")
	_, err = nomadPack.runCommand(cmd)
	if err != nil {
		pterm.DefaultBasicText.Println("Run successfully ran.")
	}
	return err
}

func (nomadPack *NomadPack) Render(workDir string, diff bool, varFiles []string, vars map[string]string, extraParams []string) error {
	params := []string{"render"}
	for _, varFile := range varFiles {
		params = append(params, "-var-file")
		params = append(params, varFile)
	}

	for key, value := range vars {
		params = append(params, "-var")
		params = append(params, key+"="+value)
	}

	params = append(params, extraParams...)
	cmd := exec.Command(nomadPack.binaryPath, params...)
	cmd.Dir = workDir

	pterm.DefaultBasicText.Println("Running Render.")
	stdout, err := nomadPack.runCommand(cmd)
	if err != nil {
		pterm.DefaultBasicText.Println("Render successfully ran.")
	}
	pterm.Println(stdout)

	return err
}

func (nomadPack *NomadPack) envForCommand() []string {
	var env = []string{}

	// Nomad credentials
	if nomadPack.nomadAddr != "" {
		env = append(env, "NOMAD_ADDR="+nomadPack.nomadAddr)
	}
	if nomadPack.nomadToken != "" {
		env = append(env, "NOMAD_TOKEN="+nomadPack.nomadToken)
	}

	env = append(env, "HOME="+os.Getenv("HOME"))
	env = append(env, "TERM="+os.Getenv("TERM"))
	env = append(env, "PATH="+os.Getenv("PATH"))

	return env
}

func (nomadPack *NomadPack) runCommand(cmd *exec.Cmd) (string, error) {
	cmd.Env = append(cmd.Env, nomadPack.envForCommand()...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		pterm.Error.Println("Error running command")
		pterm.Error.Println("Command:", cmd.String())
		pterm.Error.Println(stdout.String())
		pterm.Error.Println(stderr.String())
		return stdout.String(), err
	}

	return stdout.String(), nil
}

func (nomadPack *NomadPack) ensureValidAuth() error {
	if nomadPack.nomadAddr == "" {
		return fmt.Errorf("Nomad address is required")
	}
	if nomadPack.nomadToken == "" {
		return fmt.Errorf("Nomad token is required")
	}
	nomadClient, err := nomad.NewClient(&nomad.Config{Address: nomadPack.nomadAddr, SecretID: nomadPack.nomadToken})
	if err != nil {
		return err
	}
	_, err = nomadClient.Status().Peers()
	if err != nil {
		return fmt.Errorf("Could not connect to Nomad, error while trying to fetch then status of peers: %v", err)
	}

	return nil
}
