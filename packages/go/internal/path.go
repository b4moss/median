package internal

import (
	"errors"

	shardian "github.com/b4moss/shardian/packages/go"
)

var ErrEmptyShardName = errors.New("path: empty filename")

// BuildStoragePath runs shardian on the final filename.
// Returned path uses no leading slash (StripHeadSlash).
func BuildStoragePath(fileName string, dirLetterCount, dirNestDepth int) (string, error) {
	if fileName == "" {
		return "", ErrEmptyShardName
	}
	lc := dirLetterCount
	nd := dirNestDepth
	opt := &shardian.Option{
		DirLetterCount: &lc,
		DirNestDepth:   &nd,
		StripHeadSlash: true,
	}
	return shardian.Shardian(fileName, opt)
}
