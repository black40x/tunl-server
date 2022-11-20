package server

import (
	"github.com/black40x/gover"
)

var Version = "1.0.0"
var ClientMinimumVersion = "1.0.0"

func CheckUpdates() *gover.Version {
	currentV, _ := gover.NewVersion(Version)
	latestV, err := gover.GetGithubVersion("black40x", "tunl-server")
	if err == nil {
		ver, _ := latestV.GetVersion()
		if ver.NewerThan(*currentV) {
			return ver
		}
	}

	return nil
}

func CheckClientUp(ver string) bool {
	clientMinVer, _ := gover.NewVersion(ClientMinimumVersion)
	return clientMinVer.NewerThanStr(ver)
}
