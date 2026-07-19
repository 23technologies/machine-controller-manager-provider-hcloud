// Package version holds the build-time version of the provider binary.
package version

// Version is set at build time via -ldflags "-X .../pkg/version.Version=<v>".
var Version = "dev"
