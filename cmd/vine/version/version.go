package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitCommit = "955d212c"
	GitTag    = "v0.22.1"
	BuildDate = "1629439933"
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
	if strings.Count(v, ".") > 1 {
		v = v[:strings.LastIndex(v, ".")]
	}
	return v
}
