//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const Name = "netsh-dhcp-to"

var Default = Build

// buildInfo returns the version string stamped into binaries: the git
// describe output, a build timestamp, and the building user@host.
func buildInfo() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get build info: %v", err)
	}
	gitInfo := strings.TrimSpace(string(output))
	datetime := time.Now().Format(time.RFC1123)
	user := os.Getenv("USER")
	hostname := os.Getenv("HOSTNAME")
	return fmt.Sprintf("%s built %s by %s@%s", gitInfo, datetime, user, hostname), nil
}

// runBuild invokes go build for ./cmd/scopes with the version stamp applied.
func runBuild(name string, envVars ...string) error {
	version, err := buildInfo()
	if err != nil {
		return err
	}
	if err := os.MkdirAll("bin", 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags",
		fmt.Sprintf("-s -w -X 'main.Version=%s'", version), "-o", name, "./cmd/scopes")
	if len(envVars) > 0 {
		cmd.Env = append(os.Environ(), envVars...)
	} else {
		cmd.Env = os.Environ()
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Build compiles a native binary for the host platform into bin/.
func Build() error {
	name := "bin/scopes"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return runBuild(name)
}

// Binaries cross-compiles release binaries for all supported platforms.
// The tool is pure Go, so no cross-toolchains are required.
func Binaries() error {
	for _, combo := range []struct {
		GOOS   string
		GOARCH string
		Ext    string
	}{
		{"linux", "amd64", ""},
		{"linux", "arm64", ""},
		{"darwin", "amd64", ""},
		{"darwin", "arm64", ""},
		{"windows", "amd64", ".exe"},
	} {
		name := fmt.Sprintf("bin/scopes-%s-%s%s", combo.GOOS, combo.GOARCH, combo.Ext)
		envVars := []string{
			"CGO_ENABLED=0",
			"GOOS=" + combo.GOOS,
			"GOARCH=" + combo.GOARCH,
		}
		if err := runBuild(name, envVars...); err != nil {
			return err
		}
	}
	return nil
}

// Clean removes build outputs.
func Clean() error {
	return os.RemoveAll("bin")
}
