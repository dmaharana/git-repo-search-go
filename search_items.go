package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (c *YamlConfig) searchItemsInBatches() [][]string {
	var batchSize = c.Repository.ConcurrentSearches
	
	header := []string{"REPO_URL", "BRANCH", "SEARCH_TERM", "FILE_NAME", "LINE_NUMBER", "CONTENT"}
	outputFilePath := path.Join("./", c.Repository.OutputFile)

	log.Println("Searching repos in batches...")

	searchResults := make([][]string, 0)

	// search repos in batches of 10
	for i := 0; i < len(c.Repository.Names); i += batchSize {
		end := i + batchSize
		if end > len(c.Repository.Names) {
			end = len(c.Repository.Names)
		}

		batch := c.Repository.Names[i:end]

		// search batch in parallel
		var wg sync.WaitGroup
		wg.Add(len(batch))
		for _, repo := range batch {
			go func(r string) {
				defer wg.Done()
				rResult := searchRepo(c, r)
				if rResult != nil {
					searchResults = append(searchResults, rResult...)
				}
			}(repo)
		}
		wg.Wait()

	}

	// write results to file
	if len(searchResults) > 0 {
		log.Println("Writing search results to file...")
		searchResults = append([][]string{header}, searchResults...)

		writeCSV(searchResults, outputFilePath)
	}
	return searchResults
}

func searchRepo(c *YamlConfig, repoUrl string) [][]string {
	log.Printf("Searching repo: %s\n", repoUrl)

	repoName := repoUrl[strings.LastIndex(repoUrl, "/")+1:]
	repoName = repoName[:strings.LastIndex(repoName, ".")]
	log.Println("Repository name:", repoName)

	searchResults := make([][]string, 0)

	checkoutDir := path.Join(c.Repository.CloneDir, repoName)

	// if repo clean is set then delete the clone path
	if c.Repository.CleanUpDir {
		log.Println("Deleting clone directory: " + checkoutDir)
		os.RemoveAll(checkoutDir)
	}

	// clone repo
	r, err := git.PlainClone(checkoutDir, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Println("Error cloning repo:", err)

		r, err = git.PlainOpen(checkoutDir)
		if err != nil {
			log.Println("Error cloning repo:", err)
			return nil
		}
	}

	// checkout one branch at a time and search for terms
	bList, err := findAllBranches(r)
	if err != nil {
		log.Println("Error getting branches:", err)
		return nil
	}

	log.Println("Branches: ", bList)

	w, err := r.Worktree()
	if err != nil {
		log.Println("Error getting worktree:", err)
		return nil
	}

	for _, branch := range bList {
		log.Println("Checking out branch: ", branch)

		w.Pull(&git.PullOptions{RemoteName: "origin"})

		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(branch),
		})
		if err != nil {
			log.Println("Error checking out branch: ", err)
			continue
		}

		// search terms in branch
		results := searchTermsInBranch(c, w, repoName, branch)

		if len(results) > 0 {
			searchResults = append(searchResults, results...)
		}

	}

	log.Println("Clean up flag: " + fmt.Sprint(c.Repository.CleanUpDir))
	if c.Repository.CleanUpDir {
		// delete clone directory
		log.Println("Deleting clone directory: " + checkoutDir)
		os.RemoveAll(checkoutDir)
	} else {
		log.Println("Skipping cleanup.  Set CleanUpDir to 'true' to delete clone directory: " + checkoutDir)
	}

	return searchResults
}

func findAllBranches(r *git.Repository) ([]string, error) {
	log.Println("Branches: ")
	// branches, err := r.Branches()
	branches, err := r.References()
	if err != nil {
		// log.Fatal(err)
		return nil, nil
	}

	branchList := make([]string, 0)
	count := 0

	branches.ForEach(func(b *plumbing.Reference) error {
		if b.Type() != plumbing.HashReference {
			return nil
		}

		bname := b.Name().String()
		if strings.Contains(bname, "origin") {
			count++
			// log.Println(b.Name())
			// log.Println(bname)
			branchList = append(branchList, bname)
		}
		return nil
	})

	log.Println("Total branch(es): " + fmt.Sprint(count))

	return branchList, nil
}

func searchTermsInBranch(c *YamlConfig, w *git.Worktree, repo string, branch string) [][]string {
	// search for the terms

	results := [][]string{}

	for _, sterm := range c.Repository.SearchTerms {
		log.Println("Searching for term: " + sterm)
		// searchTerm(directory, term)
		searchPattern := sterm
		if c.Repository.MatchWord {
			log.Println("Matching whole word")
			searchPattern = "\\b" + searchPattern + "\\b"
		}

		if !c.Repository.SearchCaseSensitive {
			log.Println("Case insensitive search")
			searchPattern = "(?i)" + sterm
		}

		grepOptions := git.GrepOptions{
			Patterns: []*regexp.Regexp{regexp.MustCompile(searchPattern)},
		}
		gr, err := w.Grep(&grepOptions)
		if err != nil {
			log.Fatal(err)
		}
		// log.Println("Found: " + fmt.Sprint(gr))
		for _, g := range gr {
			log.Println(g.TreeName)
			log.Println(g.FileName)
			log.Println(g.LineNumber)
			// log.Println(g.Content)

			row := []string{repo, branch, sterm, g.FileName, fmt.Sprint(g.LineNumber), g.Content}
			results = append(results, row)
		}

	}

	return results
}
