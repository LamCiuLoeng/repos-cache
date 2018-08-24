package main

import (
	"github.com/LamCiuLoeng/repos-cache/git"
	"github.com/LamCiuLoeng/repos-cache/util"
)

func main() {
	c := util.NewConfig()
	srv := git.NewGitServer(c)
	srv.Run()
}
