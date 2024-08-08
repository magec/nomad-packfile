package nomadpack

import (
	"testing"

	"github.com/magec/nomad-packfile/test"
	"github.com/pterm/pterm"
)

func TestNomadPackNewWithInvalidPath(t *testing.T) {
	_, err := New("I_DONT_EXISTS", test.GetLogger(t))
	if err == nil {
		t.Fatal("Expected an error")
	}
}

func TestNomadPack(t *testing.T) {
	nomadPack := nomadPack(t)
	if nomadPack == nil {
		t.Fatalf("failed to create nomad pack")
	}
}

func TestNomadPackAddRegistryOk(t *testing.T) {
	nomadPack := nomadPack(t)
	main := "main"
	alertmanager := "alertmanager"
	err := nomadPack.AddRegistry("testing-community", "github.com/hashicorp/nomad-pack-community-registry", &main, &alertmanager)
	if err != nil {
		t.Fatalf("failed to add registry: %v", err)
	}
}

func TestNomadPackPlanWithoutCredentials(t *testing.T) {
	nomadPack := nomadPack(t)
	err := nomadPack.Plan("", true, "", "", "", nil, nil)
	if err == nil {
		t.Fatal("Expected error while adding registry.")
	}
}

func TestNomadPackAddRegistryFailed(t *testing.T) {
	nomadPack := nomadPack(t)
	err := nomadPack.AddRegistry("testing-community", "NO_URL", nil, nil)
	if err == nil {
		t.Fatal("Expected error while adding registry.")
	}
}

// helpers
func nomadPack(t *testing.T) *NomadPack {
	pterm.DisableOutput()
	nomadPack, err := New(test.PathForAsset(t, "bin/nomad-pack"), test.GetLogger(t))
	if err != nil {
		t.Fatalf("failed to create nomad pack: %v", err)
	}

	if nomadPack == nil {
		t.Fatalf("failed to create nomad pack")
	}

	return nomadPack
}
