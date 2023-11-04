## This utility helps in searching specific terms in the listed repositories

### Input parameters included in the yaml config file:

    - list of git repositories (note currently only public repos are supported)
    - list of search terms
    - search options - case sensitive & word only

### Finally if any results are found then a CSV file is created with the results

    - REPO_URL, BRANCH, SEARCH_TERM, FILE_NAME, LINE_NUMBER, CONTENT
    - Output file name will be accepted from the config file

`Usage: ./search-github-branches <config yaml file>`

### This project was developed with [Amazon CodeWhisperer](https://aws.amazon.com/codewhisperer/)
