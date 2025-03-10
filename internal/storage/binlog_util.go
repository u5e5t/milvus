package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/milvus-io/milvus/internal/common"
)

// ParseSegmentIDByBinlog parse segment id from binlog paths
// if path format is not expected, returns error
func ParseSegmentIDByBinlog(rootPath, path string) (UniqueID, error) {
	// check path contains rootPath as prefix
	if !strings.HasPrefix(path, rootPath) {
		return 0, fmt.Errorf("path \"%s\" does not contains rootPath \"%s\"", path, rootPath)
	}
	p := path[len(rootPath):]

	// remove leading "/"
	for strings.HasPrefix(p, "/") {
		p = p[1:]
	}

	// binlog path should consist of "[log_type]/collID/partID/segID/fieldID/fileName"
	keyStr := strings.Split(p, "/")
	if len(keyStr) == 5 {
		return 0, common.NewIgnorableError(fmt.Errorf("%s does not contains a file name", path))
	}

	if len(keyStr) == 6 {
		return strconv.ParseInt(keyStr[len(keyStr)-3], 10, 64)
	}
	return 0, fmt.Errorf("%s is not a valid binlog path", path)
}
