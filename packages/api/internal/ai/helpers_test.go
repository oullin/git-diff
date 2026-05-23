package ai

import "github.com/oullin/git-diff/internal/usercfg"

func testReader() usercfg.Reader {
	return usercfg.NewAtomicReader(usercfg.Defaults())
}
