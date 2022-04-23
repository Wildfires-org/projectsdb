package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func ParseAllProjects() error {
	// Get list of forests
	forests, err := GetForests()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"count": len(forests),
	}).Info("Got forests")

	// For each forest, get the links to the SOPA reports,
	// then parse those pages for project updates
	for i, forest := range forests {
		log.WithFields(log.Fields{
			"forest": forest.Name,
			"state":  forest.State,
		}).Info("Looking for projects")

		projectPages, err := GetSopaReportPages(forest.Url)
		if err != nil {
			log.WithFields(log.Fields{
				"forest": forest.Name,
				"state":  forest.State,
				"error":  err.Error(),
			}).Error("Issue getting list of SOPA report urls")
			continue
		}

		for _, projectPage := range projectPages {
			projects, err := getProjects(projectPage)
			if err != nil {
				log.WithFields(log.Fields{
					"forest": forest.Name,
					"state":  forest.State,
					"page":   projectPage,
					"error":  err.Error(),
				}).Error("Issue getting projects")
			}
			forests[i].Projects = append(forests[i].Projects, projects...)
		}

		log.WithFields(log.Fields{
			"forest": forest.Name,
			"state":  forest.State,
			"count":  len(forests[i].Projects),
		}).Info("Found projects")

		for j := range forests[i].Projects {
			if len(forests[i].Projects[j].Id) > 0 {
				log.WithFields(log.Fields{
					"forest":  forest.Name,
					"state":   forest.State,
					"project": forests[i].Projects[j].Name,
				}).Info("Getting documents")

				docs, err := getReportDocumentMeta(forests[i].Projects[j].Id)
				if err != nil {
					log.WithFields(log.Fields{
						"forest":  forest.Name,
						"state":   forest.State,
						"project": forests[i].Projects[j].Name,
					}).Error("Issue getting documents")
					continue
				}

				forests[i].Projects[j].ProjectDocuments = append(
					forests[i].Projects[j].ProjectDocuments,
					docs...,
				)

			}
		}

	}

	return saveProjectsJson(forests)
}

func saveProjectsJson(forests []Forest) error {
	data, err := json.Marshal(forests)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("data/forests.json", data, 0644)
}
