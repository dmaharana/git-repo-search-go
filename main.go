package main

// read yaml file with list of github repos
// clone the repos
// check out all the branches and search for terms defined in the yaml file
// report back to a csv file with headers REPO_NAME,BRANCH,FOUND_TXT,COUNT

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting...")

	if len(os.Args) != 2 {
		fmt.Println("Usage: ./search-github-branches <yaml file>")
		os.Exit(1)
	}

	config, err := NewConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(config)

	config.searchItemsInBatches()

}
