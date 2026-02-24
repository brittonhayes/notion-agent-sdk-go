package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	keychainService = "notion-ai"
	keychainAccount = "oauth"
)

// keychainAvailable reports whether an OS keychain backend is usable.
func keychainAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.LookPath("security")
		return err == nil
	case "linux":
		_, err := exec.LookPath("secret-tool")
		return err == nil
	case "windows":
		_, err := exec.LookPath("powershell.exe")
		return err == nil
	default:
		return false
	}
}

// keychainStore saves a secret string to the OS keychain.
func keychainStore(secret string) error {
	switch runtime.GOOS {
	case "darwin":
		return darwinKeychainStore(secret)
	case "linux":
		return linuxSecretStore(secret)
	case "windows":
		return windowsCredStore(secret)
	default:
		return fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// keychainLoad retrieves a secret string from the OS keychain.
func keychainLoad() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return darwinKeychainLoad()
	case "linux":
		return linuxSecretLoad()
	case "windows":
		return windowsCredLoad()
	default:
		return "", fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// keychainDelete removes the secret from the OS keychain.
func keychainDelete() error {
	switch runtime.GOOS {
	case "darwin":
		return darwinKeychainDelete()
	case "linux":
		return linuxSecretDelete()
	case "windows":
		return windowsCredDelete()
	default:
		return fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

// --- macOS Keychain (via `security` CLI) ---

func darwinKeychainStore(secret string) error {
	// -U updates if the entry already exists.
	cmd := exec.Command("security", "add-generic-password",
		"-a", keychainAccount,
		"-s", keychainService,
		"-w", secret,
		"-U",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain store: %w: %s", err, string(out))
	}
	return nil
}

func darwinKeychainLoad() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-a", keychainAccount,
		"-s", keychainService,
		"-w",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("keychain load: %w: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func darwinKeychainDelete() error {
	cmd := exec.Command("security", "delete-generic-password",
		"-a", keychainAccount,
		"-s", keychainService,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Not an error if the entry doesn't exist.
		if strings.Contains(string(out), "could not be found") {
			return nil
		}
		return fmt.Errorf("keychain delete: %w: %s", err, string(out))
	}
	return nil
}

// --- Linux Secret Service (via `secret-tool` CLI) ---

func linuxSecretStore(secret string) error {
	cmd := exec.Command("secret-tool", "store",
		"--label", "Notion AI OAuth Token",
		"service", keychainService,
		"account", keychainAccount,
	)
	cmd.Stdin = strings.NewReader(secret)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("secret-tool store: %w: %s", err, string(out))
	}
	return nil
}

func linuxSecretLoad() (string, error) {
	cmd := exec.Command("secret-tool", "lookup",
		"service", keychainService,
		"account", keychainAccount,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("secret-tool lookup: %w: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func linuxSecretDelete() error {
	cmd := exec.Command("secret-tool", "clear",
		"service", keychainService,
		"account", keychainAccount,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("secret-tool clear: %w: %s", err, string(out))
	}
	return nil
}

// --- Windows DPAPI (via PowerShell) ---
//
// Uses ConvertFrom-SecureString / ConvertTo-SecureString which encrypt with
// the Windows Data Protection API (DPAPI), scoped to the current user account.
// The encrypted token is stored at %LOCALAPPDATA%\notion-ai\token.enc.

const windowsEncFile = "token.enc"

// escapePowerShell escapes single quotes for PowerShell single-quoted strings.
func escapePowerShell(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func windowsCredStore(secret string) error {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$dir = Join-Path $env:LOCALAPPDATA '%s'
if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
$path = Join-Path $dir '%s'
$secure = ConvertTo-SecureString -String '%s' -AsPlainText -Force
$encrypted = ConvertFrom-SecureString $secure
Set-Content -Path $path -Value $encrypted -Force
`, configDirName, windowsEncFile, escapePowerShell(secret))

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dpapi store: %w: %s", err, string(out))
	}
	return nil
}

func windowsCredLoad() (string, error) {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$path = Join-Path $env:LOCALAPPDATA '%s\%s'
if (-not (Test-Path $path)) { exit 1 }
$encrypted = Get-Content $path
$secure = ConvertTo-SecureString $encrypted
$bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
$plain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
[System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
Write-Output $plain
`, configDirName, windowsEncFile)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("dpapi load: %w: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func windowsCredDelete() error {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$path = Join-Path $env:LOCALAPPDATA '%s\%s'
if (Test-Path $path) { Remove-Item $path -Force }
`, configDirName, windowsEncFile)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dpapi delete: %w: %s", err, string(out))
	}
	return nil
}
