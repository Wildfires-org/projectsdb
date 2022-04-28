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
		forests[i] = GetAllForestData(forest)
		if i > 2 {
			break
		}
	}

	return saveProjectsJson(forests)
}

func GetAllForestData(forest Forest) Forest {
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
		return forest
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
		forest.Projects = append(forest.Projects, projects...)
	}

	log.WithFields(log.Fields{
		"forest": forest.Name,
		"state":  forest.State,
		"count":  len(forest.Projects),
	}).Info("Found projects")

	for j := range forest.Projects {
		if len(forest.Projects[j].Id) > 0 {
			log.WithFields(log.Fields{
				"forest":  forest.Name,
				"state":   forest.State,
				"project": forest.Projects[j].Name,
			}).Info("Getting documents")

			docs, err := getReportDocumentMeta(forest.Projects[j].Id)
			if err != nil {
				log.WithFields(log.Fields{
					"forest":  forest.Name,
					"state":   forest.State,
					"project": forest.Projects[j].Name,
				}).Error("Issue getting documents")
				continue
			}

			forest.Projects[j].ProjectDocuments = append(
				forest.Projects[j].ProjectDocuments,
				docs...,
			)

		}
	}

	return forest
}

func saveProjectsJson(forests []Forest) error {
	data, err := json.Marshal(forests)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("data/forests.json", data, 0644)
}
