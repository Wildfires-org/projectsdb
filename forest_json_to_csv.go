package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type ForestJsonToCsvConfig struct {
	ForestDataFile string `required help:"Forest Data JSON File" type:"path"`
}

func ForestJsonToCsv(config ForestJsonToCsvConfig) error {
	// Read file off disk into memory
	file, err := ioutil.ReadFile(config.ForestDataFile)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  config.ForestDataFile,
			"error": err.Error(),
		}).Error("Unable to open forest data file")
		return err
	}

	forests := []Forest{}

	// Parse data blob into struct
	err = json.Unmarshal([]byte(file), &forests)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  config.ForestDataFile,
			"error": err.Error(),
		}).Error("Unable to parse forest data file")
		return err
	}

	filePath := "data/forest.csv"
	err = writeForestCsv(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}

	filePath = "data/projects.csv"
	err = writeProjectsCsv(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}

	filePath = "data/project_updates.csv"
	err = writeProjectUpdatesCsv(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}

	return nil
}

// TODO(hank)
func writeForestCsv(path string) error {
	return nil
}

// TODO(hank)
func writeProjectsCsv(path string) error {
	return nil
}

// TODO(hank)
func writeProjectUpdatesCsv(path string) error {
	return nil
}
