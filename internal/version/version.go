// Package version provides version information for the application.
// Variables are injected at build time via -ldflags.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version (e.g., v0.1.0)
	// Set via -ldflags "-X 'github.com/nightmare-assault/nightmare-assault/internal/version.Version=...'"
	Version = "dev"

	// Commit is the git commit hash
	// Set via -ldflags "-X 'github.com/nightmare-assault/nightmare-assault/internal/version.Commit=...'"
	Commit = "unknown"

	// BuildTime is the build timestamp
	// Set via -ldflags "-X 'github.com/nightmare-assault/nightmare-assault/internal/version.BuildTime=...'"
	BuildTime = "unknown"

	// GoVersion is the Go version used to build
	// Set via -ldflags "-X 'github.com/nightmare-assault/nightmare-assault/internal/version.GoVersion=...'"
	GoVersion = "unknown"
)

// Info holds all version information.
type Info struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
	OS        string
	Arch      string
}

// GetInfo returns the complete version information.
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// PrintVersion outputs formatted version information to stdout.
func PrintVersion() {
	fmt.Printf("Nightmare Assault %s\n", Version)
	fmt.Printf("Built: %s\n", BuildTime)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Go: %s\n", GoVersion)
}

// GetVersion returns just the version string.
func GetVersion() string {
	return Version
}

// String returns a one-line version string suitable for display.
func String() string {
	return fmt.Sprintf("Nightmare Assault %s (commit: %s)", Version, Commit)
}
