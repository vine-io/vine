package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitCommit = "d609b535"
	GitTag    = "v1.4.9"
	BuildDate = "1658111086"
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
