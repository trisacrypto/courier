package courier

import "fmt"

// Version of the current build
const (
	VersionMajor         = 0
	VersionMinor         = 1
	VersionPatch         = 0
	VersionReleaseLevel  = "beta"
	VersionReleaseNumber = 1
)

// Set the GitVersion via -ldflags="-X 'github.com/trisacrypto/courier/pkg.GitVersion=$(git rev-parse --short HEAD)'"
var GitVersion string

// Returns the semantic version for the current build.
func Version() string {
	versionCore := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	if VersionReleaseLevel != "" {
		if VersionReleaseNumber > 0 {
			versionCore = fmt.Sprintf("%s-%s.%d", versionCore, VersionReleaseLevel, VersionReleaseNumber)
		} else {
			versionCore = fmt.Sprintf("%s-%s", versionCore, VersionReleaseLevel)
		}
	}

	if GitVersion != "" {
		versionCore = fmt.Sprintf("%s (%s)", versionCore, GitVersion)
	}

	return versionCore
}
