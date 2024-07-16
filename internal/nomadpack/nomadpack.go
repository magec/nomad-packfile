package nomadpack

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"go.uber.org/zap"
)

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

func (nomadPack *NomadPack) AddRegistry(name, source string, ref, target *string) error {
	pterm.DefaultBasicText.Println("Adding registry", name, source)

	cmd := exec.Command(nomadPack.binaryPath, "registry", "add", name, source)
	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		pterm.Error.Println("Error adding registry", name, source)
		pterm.Error.Println(stderr.String())
		return err
	}

	pterm.DefaultBasicText.Println("Successfully added.")
	return nil

}

func (nomadPack *NomadPack) Plan(workDir string, diff bool, ref string, registry string, pack string, varFiles []string, vars map[string]string) error {
	args := []string{"plan", "--diff", "--exit-code-makes-changes=0"}
	if registry != "" {
		args = append(args, "--registry")
		args = append(args, registry)
	}
	if ref != "" {
		args = append(args, "--ref")
		args = append(args, ref)
	}

	for _, varFile := range varFiles {
		args = append(args, "-var-file")
		args = append(args, varFile)
	}

	for key, value := range vars {
		args = append(args, "-var")
		args = append(args, key+"="+value)
	}
	args = append(args, pack)

	cmd := exec.Command(nomadPack.binaryPath, args...)
	if nomadPack.nomadAddr != "" {
		cmd.Env = append(cmd.Env, "NOMAD_ADDR="+nomadPack.nomadAddr)
	}
	if nomadPack.nomadToken != "" {
		cmd.Env = append(cmd.Env, "NOMAD_TOKEN="+nomadPack.nomadToken)
	}

	fmt.Println("Running plan", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workDir
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "TERM="+os.Getenv("TERM"))

	if err := cmd.Run(); err != nil {
		pterm.Error.Println("Error running plan")
		pterm.Error.Println(stdout.String())
		pterm.Error.Println(stderr.String())
		return err
	}
	pterm.Println(stdout.String())

	return nil
}

func (nomadPack *NomadPack) Run(workDir string, diff bool, ref string, registry string, pack string, varFiles []string, vars map[string]string) error {
	args := []string{"run"}
	if registry != "" {
		args = append(args, "--registry")
		args = append(args, registry)
	}
	if ref != "" {
		args = append(args, "--ref")
		args = append(args, ref)
	}

	for _, varFile := range varFiles {
		args = append(args, "-var-file")
		args = append(args, varFile)
	}

	for key, value := range vars {
		args = append(args, "-var")
		args = append(args, key+"="+value)
	}
	args = append(args, pack)

	cmd := exec.Command(nomadPack.binaryPath, args...)
	if nomadPack.nomadAddr != "" {
		cmd.Env = append(cmd.Env, "NOMAD_ADDR="+nomadPack.nomadAddr)
	}
	if nomadPack.nomadToken != "" {
		cmd.Env = append(cmd.Env, "NOMAD_TOKEN="+nomadPack.nomadToken)
	}

	fmt.Println("Running plan", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workDir
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "TERM="+os.Getenv("TERM"))

	if err := cmd.Run(); err != nil {
		pterm.Error.Println("Error running plan")
		pterm.Error.Println(stdout.String())
		pterm.Error.Println(stderr.String())
		return err
	}
	pterm.Println(stdout.String())

	return nil
}

func (nomadPack *NomadPack) Render(workDir string, diff bool, ref string, registry string, pack string, varFiles []string, vars map[string]string) error {
	args := []string{"render"}
	if registry != "" {
		args = append(args, "--registry")
		args = append(args, registry)
	}
	if ref != "" {
		args = append(args, "--ref")
		args = append(args, ref)
	}

	for _, varFile := range varFiles {
		args = append(args, "-var-file")
		args = append(args, varFile)
	}

	for key, value := range vars {
		args = append(args, "-var")
		args = append(args, key+"="+value)
	}
	args = append(args, pack)

	cmd := exec.Command(nomadPack.binaryPath, args...)
	if nomadPack.nomadAddr != "" {
		cmd.Env = append(cmd.Env, "NOMAD_ADDR="+nomadPack.nomadAddr)
	}
	if nomadPack.nomadToken != "" {
		cmd.Env = append(cmd.Env, "NOMAD_TOKEN="+nomadPack.nomadToken)
	}

	fmt.Println("Running plan", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workDir
	cmd.Env = append(cmd.Env, "HOME="+os.Getenv("HOME"))
	cmd.Env = append(cmd.Env, "TERM="+os.Getenv("TERM"))

	if err := cmd.Run(); err != nil {
		pterm.Error.Println("Error running plan")
		pterm.Error.Println(stdout.String())
		pterm.Error.Println(stderr.String())
		return err
	}
	pterm.Println(stdout.String())

	return nil
}
