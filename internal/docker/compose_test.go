package docker

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNormalizeComposeConfig(t *testing.T) {
	repoPath := t.TempDir()
	raw := []byte(`{
  "services": {
    "frontend": {
      "build": {"context": ".", "dockerfile": "Dockerfile"},
      "ports": [{"target": 3000, "published": "3000", "protocol": "tcp"}],
      "labels": {"devinfra.public": "true"}
    },
    "backend": {
      "image": "node:alpine",
      "ports": [{"target": 5000, "published": "5000", "protocol": "tcp"}]
    },
    "database": {
      "image": "postgres:17",
      "volumes": [{"type": "volume", "source": "data", "target": "/var/lib/postgresql/data"}]
    }
  },
  "volumes": {"data": {"name": "old_data"}},
  "networks": {"default": {"name": "old_default"}}
}`)

	normalized, err := normalizeComposeConfig(raw, repoPath, "deployment-7", 9007)
	if err != nil {
		t.Fatal(err)
	}

	var config map[string]any
	if err := json.Unmarshal(normalized, &config); err != nil {
		t.Fatal(err)
	}
	services := config["services"].(map[string]any)
	frontend := services["frontend"].(map[string]any)
	backend := services["backend"].(map[string]any)
	ports := frontend["ports"].([]any)
	port := ports[0].(map[string]any)

	if port["published"] != "9007" {
		t.Fatalf("expected dynamic host port 9007, got %v", port["published"])
	}
	if _, exists := backend["ports"]; exists {
		t.Fatal("backend port should not be published")
	}
	if frontend["restart"] != "unless-stopped" {
		t.Fatal("restart policy was not applied")
	}
	build := frontend["build"].(map[string]any)
	if !filepath.IsAbs(build["context"].(string)) {
		t.Fatal("build context should be stored as an absolute path")
	}
	volumes := config["volumes"].(map[string]any)
	data := volumes["data"].(map[string]any)
	if data["name"] != "deployment-7_data" {
		t.Fatalf("volume was not isolated: %v", data["name"])
	}
}

func TestNormalizedConfigIsAcceptedByDockerCompose(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker CLI is not installed")
	}
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		t.Skip("docker compose plugin is not installed")
	}

	raw := []byte(`{"services":{"web":{"image":"nginx:alpine","ports":[{"target":80,"published":"80"}]}}}`)
	normalized, err := normalizeComposeConfig(raw, t.TempDir(), "deployment-1", 9001)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "compose.json")
	if err := os.WriteFile(configPath, normalized, 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("docker", "compose", "-f", configPath, "config", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("normalized config was rejected: %s: %v", output, err)
	}
}

func TestComposeCLIOutputCanBePersisted(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker CLI is not installed")
	}
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		t.Skip("docker compose plugin is not installed")
	}

	repoPath := t.TempDir()
	composeFile := filepath.Join(repoPath, "compose.yaml")
	composeYAML := []byte(`services:
  frontend:
    image: nginx:alpine
    labels:
      devinfra.public: "true"
    ports:
      - "3000:80"
  backend:
    image: node:alpine
    ports:
      - "5000:5000"
  database:
    image: postgres:17
    volumes:
      - data:/var/lib/postgresql/data
volumes:
  data:
`)
	if err := os.WriteFile(composeFile, composeYAML, 0600); err != nil {
		t.Fatal(err)
	}

	client := &Client{}
	raw, err := client.readComposeConfig(composeFile, repoPath)
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := normalizeComposeConfig(raw, repoPath, "deployment-2", 9002)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "compose.json")
	if err := os.WriteFile(configPath, normalized, 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("docker", "compose", "-f", configPath, "config", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("persisted compose CLI output was rejected: %s: %v", output, err)
	}
}

func TestNormalizeComposeConfigRejectsDangerousServices(t *testing.T) {
	repoPath := t.TempDir()
	tests := map[string]string{
		"privileged":   `{"services":{"web":{"image":"nginx","privileged":true,"ports":[{"target":80,"published":"80"}]}}}`,
		"bind mount":   `{"services":{"web":{"image":"nginx","ports":[{"target":80,"published":"80"}],"volumes":[{"type":"bind","source":"/","target":"/host"}]}}}`,
		"host network": `{"services":{"web":{"image":"nginx","network_mode":"host","ports":[{"target":80,"published":"80"}]}}}`,
	}

	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := normalizeComposeConfig([]byte(raw), repoPath, "deployment-1", 9001); err == nil {
				t.Fatal("expected unsafe compose configuration to be rejected")
			}
		})
	}
}

func TestNormalizeComposeConfigRejectsEscapingBuildContext(t *testing.T) {
	repoPath := t.TempDir()
	outside := filepath.Join(repoPath, "..", "outside")
	raw := []byte(`{"services":{"web":{"build":{"context":"` + filepath.ToSlash(outside) + `"},"ports":[{"target":80,"published":"80"}]}}}`)

	if _, err := normalizeComposeConfig(raw, repoPath, "deployment-1", 9001); err == nil {
		t.Fatal("expected escaping build context to be rejected")
	}
}
