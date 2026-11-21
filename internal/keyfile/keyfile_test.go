package keyfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializePreservesKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key")
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0600)
	first, err := os.ReadFile(path)
	if err != nil || len(first) != 65 {
		t.Fatal("missing generated key", err)
	}
	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Fatal("initialization replaced the key")
	}
}

func TestInitializeRejectsInvalidExistingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(path, []byte("incomplete"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Initialize(path); err == nil {
		t.Fatal("accepted invalid key")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "incomplete" {
		t.Fatal("modified invalid existing key")
	}
}
