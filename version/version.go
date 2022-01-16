// Package version provides the version information of the current build of
// req.
package version

// Build is the build version of req. This will either be a git SHA or a
// version number. This is set via -ldflags at build time.
var Build string
