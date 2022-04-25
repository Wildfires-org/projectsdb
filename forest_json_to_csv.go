package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
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
	err = writeForestCsv(forests, filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}

	filePath = "data/projects.csv"
	err = writeProjectsCsv(forests, filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}

	filePath = "data/project_updates.csv"
	err = writeProjectUpdatesCsv(forests, filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}

	return nil
}

func writeForestCsv(forests []Forest, path string) error {
	// open file
	csvFile, err := os.Create(path)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
		return err
	}
	defer csvFile.Close()
	// setup columns
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{
		"Forest Name",
		"Forest State",
		"Forest URL",
		"Forest ID",
	})
	// build rows
	for _, forest := range forests {
		rows := [][]string{}
		rows = append(rows, append([]string{
			forest.Name,
			forest.State,
			forest.Url,
			fmt.Sprint(forest.Id),
		}, []string{}...))

		// write rows to file
		writer.WriteAll(rows)
	}
	writer.Flush()
	return nil
}

func writeProjectsCsv(forests []Forest, path string) error {
	// open file
	csvFile, err := os.Create(path)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
		return err
	}
	defer csvFile.Close()

	// build rows
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{
		"Forest Name",
		"Forest State",
		"Forest URL",
		"Forest ID",
		"Project Name",
		"Project ID",
		"Project Purposes",
		"Project Status",
		"Project Decision",
		"Project Expected Implementation",
		"Project Contact Name",
		"Project Contact Email",
		"Project Contact Phone",
		"Project Description",
		"Project Web Link",
		"Project Location",
		"Project Region",
		"Project District",
		"Project SOPA Report Date",
		"Project Code",
		"Project Documents",
	})

	// remove all but one project update
	for i, _ := range forests {
		projectMap := make(map[string]ProjectUpdate)
		for j, _ := range forests[i].Projects {
			currentProjectUpdate := forests[i].Projects[j]
			// check if exists in map
			if mostRecentProjectUpdate, ok := projectMap[currentProjectUpdate.Name]; ok {
				mostRecentDate, _ := time.Parse("2006-01", mostRecentProjectUpdate.SopaReportDate)
				currentProjectDate, _ := time.Parse("2006-01", currentProjectUpdate.SopaReportDate)
				if mostRecentDate.Before(currentProjectDate) {
					fmt.Printf("updating most recent project\n")
					projectMap[currentProjectUpdate.Name] = currentProjectUpdate // found a more recent update
				}
			} else {
				// if not already in map add project
				projectMap[currentProjectUpdate.Name] = currentProjectUpdate
			}

		}
		forests[i].Projects = nil
		for _, update := range projectMap {
			update.cleanDescription()
			forests[i].Projects = append(forests[i].Projects, update)
		}
	}

	//write rows to file
	for _, forest := range forests {
		writer.WriteAll(forest.AsCsv())
	}
	writer.Flush()
	return nil
}

func writeProjectUpdatesCsv(forests []Forest, path string) error {
	// open file
	csvFile, err := os.Create(path)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
		return err
	}
	defer csvFile.Close()

	// build rows
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{
		"Forest Name",
		"Forest State",
		"Forest URL",
		"Forest ID",
		"Project Name",
		"Project ID",
		"Project Purposes",
		"Project Status",
		"Project Decision",
		"Project Expected Implementation",
		"Project Contact Name",
		"Project Contact Email",
		"Project Contact Phone",
		"Project Description",
		"Project Web Link",
		"Project Location",
		"Project Region",
		"Project District",
		"Project SOPA Report Date",
		"Project Code",
		"Project Documents",
	})

	// clean up description
	for i, _ := range forests {
		for j, _ := range forests[i].Projects {
			forests[i].Projects[j].cleanDescription()
		}
	}

	//write rows to file
	for _, forest := range forests {
		writer.WriteAll(forest.AsCsv())
	}
	writer.Flush()
	return nil
}
