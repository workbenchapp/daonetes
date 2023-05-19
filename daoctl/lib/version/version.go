package version

import (
	"runtime/debug"
	"strings"
)

var versionString = "(devel)" // to match the default from runtime/debug
var buildRevision = "(devel)"
var builtBy = ""
var buildDate = ""

func GetVersionString() string {
	// don't overwrite the version if it was set by -ldflags=-X
	if versionString != "(devel)" {
		return versionString
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		mod := &info.Main
		if mod.Replace != nil {
			mod = mod.Replace
		}
		versionString = mod.Version
	}
	return versionString
}

func GetBuildRevision() string {
	// don't overwrite the version if it was set by -ldflags=-X
	if buildRevision != "(devel)" {
		return buildRevision
	}

	modified := false

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs.revision") {
				buildRevision = setting.Value
			}
			if strings.HasPrefix(setting.Key, "vcs.modified") {
				modified = true
			}
		}
	}
	if modified {
		buildRevision = buildRevision + "-modified"
	}
	return buildRevision
}

func GetBuildDate() string {
	if buildDate != "" {
		return buildDate
	}

	// TODO: this is not the build time, its the commit time
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs.time") {
				return setting.Value
			}
		}
	}
	return ""
}
