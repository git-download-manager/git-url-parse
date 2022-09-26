package main

import (
	"fmt"
	"github.com/git-download-manager/git-url-parse"
) 

func main() {
	// args
	tempDir := "" // only use gitdownloadmanager service
	ssid := ""    // only use gitdownloadmanager service for session
	rawUrl := "https://github.com/cli/cli"
	branch := "" // if has a multi slash branch name, you can set it

	// create new repository obj
	gitRepository := gitrepository.NewGitRepository(tempDir, ssid, rawUrl, branch)

	// parse if you can
	sub := ""                  // use only breadcrumb query for find root folder
	direction := gitrepository.DirectionNone // use only breadcrumb query for which direction to go
	filename := ""             // use file user download for only one file
	if err := gitRepository.Parse(sub, direction, filename); err != nil {
		fmt.Printf("GitRepository.Parse() error = %#v", err)
		panic(err)
	}

	fmt.Printf("GitRepository.Parse() success = %#v", gitRepository)
}
