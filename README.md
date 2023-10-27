## This utility helps in searching specific terms in the listed repositories

### Input parameters include

    - list of git repositories (note currently only public repos are supported)
    - list of search terms
    - search options - case sensitive & word only

### Finally if any results are found then a CSV file is created with the results

    - REPO_URL, BRANCH, SEARCH_TERM, LINE_NUMBER, CONTENT
    - Output file name will be accepted from the config file

`Usage: ./search-github-branches <config yaml file>`

#### TODO: Perform clone and search parallely
