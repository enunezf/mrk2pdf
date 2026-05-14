package pdf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// errNoBrowser is returned when no Chromium-based browser can be located.
var errNoBrowser = errors.New(
	"No se encontró un navegador basado en Chromium. " +
		"Por favor instala Chrome o Edge, o define la variable de entorno BROWSER_PATH",
)

// FindBrowser locates a Chromium-based browser usable by chromedp.
//
// Lookup order (first match wins):
//  1. BROWSER_PATH environment variable
//  2. Standard $PATH lookup of known binary names
//  3. Canonical install paths for the current OS (and WSL, on Linux)
func FindBrowser() (string, error) {
	if env := os.Getenv("BROWSER_PATH"); env != "" {
		if info, err := os.Stat(env); err == nil && !info.IsDir() {
			return env, nil
		}
		return "", fmt.Errorf("BROWSER_PATH=%q no apunta a un ejecutable existente", env)
	}

	for _, name := range pathCandidates() {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	for _, p := range fixedCandidates() {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, nil
		}
	}

	return "", errNoBrowser
}

// pathCandidates returns binary names commonly resolvable via $PATH on at
// least one supported OS. On Windows, exec.LookPath honours PATHEXT, so
// bare "chrome" still resolves to chrome.exe when chrome is on PATH.
func pathCandidates() []string {
	return []string{
		// Linux convention (and Homebrew casks on macOS)
		"google-chrome",
		"google-chrome-stable",
		"chromium",
		"chromium-browser",
		"microsoft-edge",
		"microsoft-edge-stable",
		"brave-browser",
		// Windows / generic
		"chrome",
		"msedge",
		"brave",
	}
}

// fixedCandidates returns canonical install paths for the current OS.
// On Linux it returns nothing unless running under WSL.
func fixedCandidates() []string {
	switch runtime.GOOS {
	case "darwin":
		return darwinFixedCandidates()
	case "windows":
		return windowsFixedCandidates()
	case "linux":
		if isWSL() {
			return wslFixedCandidates()
		}
	}
	return nil
}

func darwinFixedCandidates() []string {
	return []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Google Chrome Beta.app/Contents/MacOS/Google Chrome Beta",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
	}
}

func windowsFixedCandidates() []string {
	paths := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
		`C:\Program Files (x86)\BraveSoftware\Brave-Browser\Application\brave.exe`,
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		paths = append(paths,
			filepath.Join(localAppData, `Google\Chrome\Application\chrome.exe`),
			filepath.Join(localAppData, `Microsoft\Edge\Application\msedge.exe`),
			filepath.Join(localAppData, `BraveSoftware\Brave-Browser\Application\brave.exe`),
		)
	}
	return paths
}

func wslFixedCandidates() []string {
	return []string{
		`/mnt/c/Program Files/Google/Chrome/Application/chrome.exe`,
		`/mnt/c/Program Files (x86)/Google/Chrome/Application/chrome.exe`,
		`/mnt/c/Program Files/Microsoft/Edge/Application/msedge.exe`,
		`/mnt/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe`,
		`/mnt/c/Program Files/BraveSoftware/Brave-Browser/Application/brave.exe`,
	}
}

func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// IsWindowsExecutable reports whether the given path looks like a Windows
// binary launched from WSL (lives under /mnt/<drive>/ and ends in .exe).
// chromedp needs special handling when launching such binaries since the
// .exe runs in the Windows process namespace, not WSL2's.
func IsWindowsExecutable(path string) bool {
	return strings.HasPrefix(path, "/mnt/") && strings.HasSuffix(strings.ToLower(path), ".exe")
}
