package browser

import (
	"os"
	"os/exec"
)

// FindChromiumExecutable searches for a Chromium-based browser executable.
// It first checks the RENDEREIRO_BROWSER_PATH environment variable,
// then common binary names on Linux using exec.LookPath.
func FindChromiumExecutable() string {
	// 1. Check environment variable
	if path := os.Getenv("RENDEREIRO_BROWSER_PATH"); path != "" {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	// 2. Check common Linux binary names
	candidateBinaries := []string{
		"google-chrome",
		"chromium",
		"chromium-browser",
		"vivaldi-stable",
		"brave-browser",
		"brave", // Added to the end
	}

	for _, bin := range candidateBinaries {
		if path, err := exec.LookPath(bin); err == nil {
			return path
		}
	}

	return "" // Not found
}
