package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var composeFiles = []string{
	"compose.yaml",
	"compose.yml",
	"docker-compose.yaml",
	"docker-compose.yml",
}

func FindComposeFile(repoPath string) (string, bool) {
	for _, name := range composeFiles {
		path := filepath.Join(repoPath, name)
		info, err := os.Lstat(path)
		if err == nil && info.Mode().IsRegular() {
			return path, true
		}
	}
	return "", false
}

func ComposeProjectName(deploymentID int32) string {
	return fmt.Sprintf("deployment-%d", deploymentID)
}

func ComposeConfigPath(deploymentID int32) string {
	return filepath.Join("tmp", fmt.Sprintf("deployment-%d.compose.json", deploymentID))
}

func (c *Client) DeployCompose(
	composeFile string,
	repoPath string,
	projectName string,
	configPath string,
	hostPort int,
	environment map[string]string,
) error {
	rawConfig, err := c.readComposeConfig(composeFile, repoPath, environment)
	if err != nil {
		return err
	}

	normalized, err := normalizeComposeConfig(rawConfig, repoPath, projectName, hostPort)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, normalized, 0600); err != nil {
		return err
	}

	_, err = c.runCompose(15*time.Minute, configPath, projectName, "up", "-d", "--build", "--remove-orphans")
	if err != nil {
		_, _ = c.runCompose(time.Minute, configPath, projectName, "down", "--remove-orphans")
		_ = os.Remove(configPath)
	}
	return err
}

func (c *Client) ComposeStop(configPath string, projectName string) ([]byte, error) {
	return c.runCompose(time.Minute, configPath, projectName, "stop")
}

func (c *Client) ComposeStart(configPath string, projectName string) ([]byte, error) {
	return c.runCompose(time.Minute, configPath, projectName, "start")
}

func (c *Client) ComposeLogs(configPath string, projectName string) ([]byte, error) {
	return c.runCompose(15*time.Second, configPath, projectName, "logs", "--no-color", "--tail", "200")
}

func (c *Client) ComposeRemove(configPath string, projectName string) ([]byte, error) {
	return c.runCompose(time.Minute, configPath, projectName, "down", "--remove-orphans", "--rmi", "local", "--volumes")
}

func (c *Client) ComposeIsRunning(configPath string, projectName string) (bool, error) {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	output, err := c.runCompose(15*time.Second, configPath, projectName, "ps", "--status", "running", "--quiet")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func (c *Client) runCompose(
	timeout time.Duration,
	configPath string,
	projectName string,
	args ...string,
) ([]byte, error) {
	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	commandArgs := []string{"compose", "-f", absConfig, "-p", projectName}
	commandArgs = append(commandArgs, args...)
	output, err := exec.CommandContext(ctx, "docker", commandArgs...).CombinedOutput()
	if err != nil {
		return output, commandError("docker compose", output, err)
	}
	return output, nil
}

func (c *Client) readComposeConfig(composeFile string, repoPath string, environment map[string]string) ([]byte, error) {
	absFile, err := filepath.Abs(composeFile)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	args := []string{
		"compose",
		"-f",
		absFile,
		"-p",
		"devinfra-validation",
		"config",
		"--format",
		"json",
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = repoPath
	cmd.Env = composeCommandEnvironment()
	for name, value := range environment {
		cmd.Env = append(cmd.Env, name+"="+value)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, commandError("docker compose config", stderr.Bytes(), err)
	}
	return output, nil
}

func composeCommandEnvironment() []string {
	allowed := []string{
		"PATH", "HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA",
		"SYSTEMROOT", "TEMP", "TMP", "TMPDIR", "COMSPEC", "PATHEXT", "WINDIR",
		"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432", "ProgramData",
		"DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH",
	}
	environment := make([]string, 0, len(allowed))
	for _, name := range allowed {
		if value, ok := os.LookupEnv(name); ok {
			environment = append(environment, name+"="+value)
		}
	}
	return environment
}

func normalizeComposeConfig(
	raw []byte,
	repoPath string,
	projectName string,
	hostPort int,
) ([]byte, error) {
	if hostPort < 1024 || hostPort > 65535 {
		return nil, errors.New("compose host port is out of range")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var config map[string]any
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("invalid compose configuration: %w", err)
	}

	services, ok := config["services"].(map[string]any)
	if !ok || len(services) == 0 {
		return nil, errors.New("compose file has no services")
	}
	if len(services) > 4 {
		return nil, errors.New("compose deployments support at most 4 services")
	}

	if err := sanitizeTopLevelResources(config, "volumes", projectName); err != nil {
		return nil, err
	}
	if err := sanitizeTopLevelResources(config, "networks", projectName); err != nil {
		return nil, err
	}
	if hasEntries(config["secrets"]) || hasEntries(config["configs"]) {
		return nil, errors.New("compose secrets and configs are not supported")
	}

	config["name"] = projectName
	serviceNames := make([]string, 0, len(services))
	for name := range services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)

	memoryMB := 768 / len(services)
	if memoryMB > 512 {
		memoryMB = 512
	}
	if memoryMB < 192 {
		memoryMB = 192
	}

	for _, name := range serviceNames {
		service, ok := services[name].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("compose service %s is invalid", name)
		}
		if err := validateComposeService(name, service, repoPath); err != nil {
			return nil, err
		}
		service["restart"] = "unless-stopped"
		service["mem_limit"] = fmt.Sprintf("%dm", memoryMB)
		service["cpus"] = json.Number("0.50")
		service["pids_limit"] = json.Number("100")
	}

	publicService, err := selectPublicService(services, serviceNames)
	if err != nil {
		return nil, err
	}

	for _, name := range serviceNames {
		service := services[name].(map[string]any)
		if name != publicService {
			delete(service, "ports")
			continue
		}

		port, err := selectPublicPort(service)
		if err != nil {
			return nil, fmt.Errorf("compose service %s: %w", name, err)
		}
		port["published"] = strconv.Itoa(hostPort)
		port["host_ip"] = "0.0.0.0"
		service["ports"] = []any{port}
	}

	return json.MarshalIndent(config, "", "  ")
}

func validateComposeService(name string, service map[string]any, repoPath string) error {
	if value, _ := service["privileged"].(bool); value {
		return fmt.Errorf("compose service %s cannot use privileged mode", name)
	}
	if value, _ := service["container_name"].(string); value != "" {
		return fmt.Errorf("compose service %s cannot set container_name", name)
	}

	for _, field := range []string{"network_mode", "pid", "ipc", "userns_mode"} {
		if value, _ := service[field].(string); strings.EqualFold(value, "host") {
			return fmt.Errorf("compose service %s cannot use host %s", name, field)
		}
	}
	for _, field := range []string{"devices", "cap_add", "security_opt", "sysctls", "develop"} {
		if hasEntries(service[field]) {
			return fmt.Errorf("compose service %s cannot set %s", name, field)
		}
	}

	if volumes, ok := service["volumes"].([]any); ok {
		for _, rawVolume := range volumes {
			volume, ok := rawVolume.(map[string]any)
			if !ok {
				return fmt.Errorf("compose service %s has an invalid volume", name)
			}
			volumeType, _ := volume["type"].(string)
			source, _ := volume["source"].(string)
			target, _ := volume["target"].(string)
			if volumeType == "bind" || strings.Contains(source, "docker.sock") || strings.Contains(target, "docker.sock") {
				return fmt.Errorf("compose service %s cannot use host bind mounts", name)
			}
		}
	}

	build, ok := service["build"].(map[string]any)
	if !ok {
		return nil
	}
	if hasEntries(build["additional_contexts"]) {
		return fmt.Errorf("compose service %s cannot use additional build contexts", name)
	}

	contextPath, _ := build["context"].(string)
	if contextPath == "" {
		return fmt.Errorf("compose service %s has no build context", name)
	}
	if err := ensurePathInside(repoPath, contextPath); err != nil {
		return fmt.Errorf("compose service %s build context: %w", name, err)
	}
	if !filepath.IsAbs(contextPath) {
		contextPath = filepath.Join(repoPath, contextPath)
	}
	contextPath, err := filepath.Abs(contextPath)
	if err != nil {
		return fmt.Errorf("compose service %s build context: %w", name, err)
	}
	build["context"] = contextPath

	if dockerfile, _ := build["dockerfile"].(string); dockerfile != "" {
		if !filepath.IsAbs(dockerfile) {
			dockerfile = filepath.Join(contextPath, dockerfile)
		}
		if err := ensurePathInside(repoPath, dockerfile); err != nil {
			return fmt.Errorf("compose service %s dockerfile: %w", name, err)
		}
	}
	return nil
}

func selectPublicService(services map[string]any, names []string) (string, error) {
	var labeled []string
	var published []string
	for _, name := range names {
		service := services[name].(map[string]any)
		ports, _ := service["ports"].([]any)
		if len(ports) == 0 {
			continue
		}
		published = append(published, name)
		if strings.EqualFold(serviceLabel(service, "devinfra.public"), "true") {
			labeled = append(labeled, name)
		}
	}

	if len(labeled) == 1 {
		return labeled[0], nil
	}
	if len(labeled) > 1 {
		return "", errors.New("only one compose service can have devinfra.public=true")
	}

	for _, preferred := range []string{"frontend", "web", "app", "client"} {
		for _, name := range published {
			if strings.EqualFold(name, preferred) {
				return name, nil
			}
		}
	}
	if len(published) == 1 {
		return published[0], nil
	}
	return "", errors.New("compose file must publish one service or label one service with devinfra.public=true")
}

func selectPublicPort(service map[string]any) (map[string]any, error) {
	ports, _ := service["ports"].([]any)
	if len(ports) == 0 {
		return nil, errors.New("public service has no published port")
	}

	wanted := serviceLabel(service, "devinfra.port")
	for _, rawPort := range ports {
		port, ok := rawPort.(map[string]any)
		if !ok {
			continue
		}
		target := fmt.Sprint(port["target"])
		if wanted == "" || target == wanted {
			return port, nil
		}
	}
	return nil, fmt.Errorf("published port %s was not found", wanted)
}

func sanitizeTopLevelResources(config map[string]any, field string, projectName string) error {
	resources, ok := config[field].(map[string]any)
	if !ok {
		return nil
	}
	for key, rawResource := range resources {
		resource, ok := rawResource.(map[string]any)
		if !ok {
			continue
		}
		if external, _ := resource["external"].(bool); external {
			return fmt.Errorf("external compose %s are not supported", field)
		}
		resource["name"] = projectName + "_" + key
	}
	return nil
}

func serviceLabel(service map[string]any, name string) string {
	labels, ok := service["labels"].(map[string]any)
	if !ok {
		return ""
	}
	value, _ := labels[name].(string)
	return value
}

func ensurePathInside(repoPath string, candidate string) error {
	base, err := filepath.Abs(repoPath)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(base, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(base, candidate)
	if err != nil {
		return err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("path escapes the repository")
	}
	return nil
}

func hasEntries(value any) bool {
	switch typed := value.(type) {
	case []any:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	case string:
		return typed != ""
	default:
		return value != nil
	}
}

func commandError(action string, output []byte, err error) error {
	message := strings.TrimSpace(string(output))
	if len(message) > 4096 {
		message = message[len(message)-4096:]
	}
	if message == "" {
		return fmt.Errorf("%s failed: %w", action, err)
	}
	return fmt.Errorf("%s failed: %s: %w", action, message, err)
}
