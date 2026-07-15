package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const maxValueLength = 16 << 10

var ErrInvalidVariable = errors.New("invalid environment variable")

func ParseEncryptionKey(value string) ([]byte, error) {
	key, err := hex.DecodeString(strings.TrimSpace(value))
	if err != nil || len(key) != 32 {
		return nil, errors.New("ENV_ENCRYPTION_KEY must be 64 hexadecimal characters")
	}
	return key, nil
}

func ValidateVariable(name string, value string) error {
	if len(name) == 0 || len(name) > 128 || !isVariableStart(name[0]) {
		return ErrInvalidVariable
	}
	for i := 1; i < len(name); i++ {
		if !isVariableCharacter(name[i]) {
			return ErrInvalidVariable
		}
	}
	if isReservedVariable(name) {
		return ErrInvalidVariable
	}
	if len(value) > maxValueLength || strings.ContainsAny(value, "\r\n\x00") {
		return ErrInvalidVariable
	}
	return nil
}

func Encrypt(key []byte, value string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(value), nil), nil
}

func Decrypt(key []byte, encrypted []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	if len(encrypted) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted environment value")
	}

	nonce := encrypted[:gcm.NonceSize()]
	value, err := gcm.Open(nil, nonce, encrypted[gcm.NonceSize():], nil)
	if err != nil {
		return "", errors.New("failed to decrypt environment value")
	}
	return string(value), nil
}

func WriteEnvFile(path string, variables map[string]string) error {
	names := make([]string, 0, len(variables))
	for name, value := range variables {
		if err := ValidateVariable(name, value); err != nil {
			return err
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var content strings.Builder
	for _, name := range names {
		fmt.Fprintf(&content, "%s=%s\n", name, variables[name])
	}
	if err := os.WriteFile(path, []byte(content.String()), 0600); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func isVariableStart(value byte) bool {
	return value == '_' || value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func isVariableCharacter(value byte) bool {
	return isVariableStart(value) || value >= '0' && value <= '9'
}

func isReservedVariable(name string) bool {
	switch name {
	case "PATH", "HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "SYSTEMROOT", "TEMP", "TMP", "TMPDIR", "COMSPEC", "PATHEXT", "WINDIR":
		return true
	}
	return strings.HasPrefix(name, "DOCKER_") || strings.HasPrefix(name, "COMPOSE_")
}
