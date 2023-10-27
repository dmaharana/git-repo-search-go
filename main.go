package main

// read yaml file with list of github repos
// clone the repos
// check out all the branches and search for terms defined in the yaml file
// report back to a csv file with headers REPO_NAME,BRANCH,FOUND_TXT,COUNT

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v2"
)

type Repo struct {
	Name     string
	Branches []string
}

type YamlConfig struct {
	Repository struct {
		Names       []string `yaml:"names"`
		SearchTerms []string `yaml:"searchTerms"`
		SearchCaseSensitive bool     `yaml:"searchCaseSensitive"`
		MatchWord bool     `yaml:"matchWord"`
		CloneDir    string   `yaml:"cloneDir"`
		CleanUpDir  bool     `yaml:"cleanUpDir"`
	}
}

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

	config.searchGitHubBranches()
}

func (c *YamlConfig) searchGitHubBranches() {
	// search for the terms
	// report back to a csv file with headers REPO_NAME,BRANCH,FOUND_TXT,COUNT
	log.Println("Searching for terms in branches...")

	for _, repo := range c.Repository.Names {
		log.Println("Repository URL: " + repo)

		// get repository name from the url
		repoName := repo[strings.LastIndex(repo, "/")+1:]
		repoName = repoName[:strings.LastIndex(repoName, ".")]
		log.Println("Repository name: " + repoName)

		// join directory and repo_name to form clone directory
		directory := path.Join(c.Repository.CloneDir, repoName)
		log.Println("Clone directory: " + directory)

		// clone the repo
		r, err := git.PlainClone(directory, false, &git.CloneOptions{
			URL:      repo,
			Progress: os.Stdout,
		})
		if err != nil {
			// log.Fatal(err)
			log.Println(err)
			// try to access the repository
			r, err = git.PlainOpen(directory)
			if err != nil {
				log.Fatal(err)
			}
			// r.Branches()
		}

		// get all the branches
		branchList, err := findAllBranches(r)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Branches: " + fmt.Sprint(branchList))

		w, err := r.Worktree()
		if err != nil {
			log.Fatal(err)
		}

		// checkout one branch at a time and search for the terms
		for _, branch := range branchList {
			log.Println("Checking out branch: " + branch)
			// checkout the branch
			w.Checkout(&git.CheckoutOptions{
				Branch: plumbing.ReferenceName(branch),
			})

			// search for the terms
			for _, sterm := range c.Repository.SearchTerms {
				log.Println("Searching for term: " + sterm)
				// searchTerm(directory, term)
				searchPattern := sterm
				if c.Repository.MatchWord {
					log.Println("Matching whole word")
					searchPattern = "\\b"+searchPattern+"\\b"
				}

				if !c.Repository.SearchCaseSensitive {
					log.Println("Case insensitive search")
					searchPattern = "(?i)"+sterm
				}


				grepOptions := git.GrepOptions{
					Patterns: []*regexp.Regexp{regexp.MustCompile(searchPattern)},
				}
				gr, err := w.Grep(&grepOptions)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Found: " + fmt.Sprint(gr))
				for _, g := range gr {
					log.Println(g.FileName)
					log.Println(g.LineNumber)
					log.Println(g.Content)
					log.Println(g.TreeName)
				}
				
			}
		}


		log.Println("Clean up flag: " + fmt.Sprint(c.Repository.CleanUpDir))
		if c.Repository.CleanUpDir {
			// delete clone directory
			log.Println("Deleting clone directory: " + directory)
			os.RemoveAll(directory)
		} else {
			log.Println("Skipping cleanup.  Set CleanUpDir to 'true' to delete clone directory: " + directory)
		}
	}
}

func findAllBranches(r *git.Repository) ([]string, error) {
	log.Println("Branches: ")
	// branches, err := r.Branches()
	branches, err := r.References()
	if err != nil {
		// log.Fatal(err)
		return nil, err
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

// NewConfig returns a new decoded Config struct
func NewConfig(path string) (*YamlConfig, error) {

	log.Println("Reading config file: " + path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create config structure
	cfg := &YamlConfig{}

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
