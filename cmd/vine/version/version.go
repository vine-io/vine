package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitCommit string
	GitTag    string
	BuildDate string
)

func Version() string {
	var vineVersion string

	if GitTag != "" {
		vineVersion = GitTag
	}

	if GitCommit != "" {
		vineVersion += fmt.Sprintf("-%s", GitCommit)
	}

	if BuildDate != "" {
		vineVersion += fmt.Sprintf("-%s", BuildDate)
	}

	if vineVersion == "" {
		vineVersion = "latest"
	}

	return vineVersion
}

func GoV() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	v = v[:strings.LastIndex(v, ".")]
	return v
}
