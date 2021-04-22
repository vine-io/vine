package tool

import (
	"testing"
)

func TestNewMod(t *testing.T) {
	c, err := New("mod.toml")
	if err != nil {
		t.Fatal(err)
	}

	if c.Package.Kind != "cluster" {
		t.Fatal("valid package kind")
	}

	if len(c.Mod) == 0 {
		t.Fatal("no module")
	}

	t.Log(c)
}

func TestNewPkg(t *testing.T) {
	c, err := New("pkg.toml")
	if err != nil {
		t.Fatal(err)
	}

	if c.Package.Kind != "single" {
		t.Fatal("valid package kind")
	}

	if len(c.Pkg.Name) == 0 {
		t.Fatal("no pkg")
	}

	t.Log(c)
}