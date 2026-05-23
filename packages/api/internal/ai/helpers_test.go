package ai

import "github.com/gocanto/git-diff/internal/usercfg"

func testReader() usercfg.Reader {
	return usercfg.NewAtomicReader(usercfg.Defaults())
}
