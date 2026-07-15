package secrets

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	encrypted, err := Encrypt(key, "secret-value")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encrypted, []byte("secret-value")) {
		t.Fatal("encrypted value contains plaintext")
	}

	value, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if value != "secret-value" {
		t.Fatalf("got %q", value)
	}
}

func TestDecryptRejectsModifiedValue(t *testing.T) {
	key := bytes.Repeat([]byte{2}, 32)
	encrypted, err := Encrypt(key, "secret-value")
	if err != nil {
		t.Fatal(err)
	}
	encrypted[len(encrypted)-1] ^= 1
	if _, err := Decrypt(key, encrypted); err == nil {
		t.Fatal("expected modified ciphertext to fail")
	}
}

func TestValidateVariable(t *testing.T) {
	for _, name := range []string{"GEMINI_API_KEY", "_PORT", "VALUE1"} {
		if err := ValidateVariable(name, "value"); err != nil {
			t.Fatalf("expected %s to be valid", name)
		}
	}
	for _, name := range []string{"", "1VALUE", "BAD-NAME", "DOCKER_HOST", "COMPOSE_FILE", "PATH"} {
		if err := ValidateVariable(name, "value"); err == nil {
			t.Fatalf("expected %s to be invalid", name)
		}
	}
	if err := ValidateVariable("VALUE", "line one\nline two"); err == nil {
		t.Fatal("expected multiline value to be invalid")
	}
}

func TestWriteEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deployment.env")
	if err := WriteEnvFile(path, map[string]string{"SECOND": "2", "FIRST": "1"}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "FIRST=1\nSECOND=2\n" {
		t.Fatalf("unexpected content: %q", content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("unexpected permissions: %o", info.Mode().Perm())
	}
}
