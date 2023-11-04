package main

import (
	"encoding/csv"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type YamlConfig struct {
	Repository struct {
		Names               []string `yaml:"names"`
		SearchTerms         []string `yaml:"searchTerms"`
		SearchCaseSensitive bool     `yaml:"searchCaseSensitive"`
		MatchWord           bool     `yaml:"matchWord"`
		CloneDir            string   `yaml:"cloneDir"`
		CleanUpDir          bool     `yaml:"cleanUpDir"`
		OutputFile          string   `yaml:"outputFile"`
	}
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


// write results to a csv file
func writeCSV(results [][]string, filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// initialize csv writer
	w := csv.NewWriter(file)
	defer w.Flush()

	// write all rows at once
	w.WriteAll(results)

	log.Println("Results written to: " + filename)

	return nil
}
