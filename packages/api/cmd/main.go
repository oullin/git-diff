package main

import (
	"os"

	"github.com/oullin/git-diff/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
