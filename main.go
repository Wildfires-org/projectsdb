package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	ParseUpdates     ParseUpdatesConfig    `cmd help:"Pull current known data from file and parse for updates. If updates..."`
	ParseAllProjects struct{}              `cmd help:"Parse all projects available and save to JSON"`
	UploadDocuments  UploadDocumentsConfig `cmd help:"TODO"`
	ForestJsonToCsv  ForestJsonToCsvConfig `cmd help:"Convert JSON blob into projects, project_updates, and Forests"`
	PalsToCsv        PALSToCsvConfig       `cmd help: "Convert PALS csv into airtable readble csv"`
	Quick            struct{}              `cmd`
}

func main() {
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "parse-updates":
		ParseUpdates(cli.ParseUpdates)

	case "parse-all-projects":
		ParseAllProjects()

	case "upload-documents":
		UploadDocuments(cli.UploadDocuments)

	case "forest-json-to-csv":
		ForestJsonToCsv(cli.ForestJsonToCsv)

	case "pals-to-csv":
		PalsToCsv(cli.PalsToCsv)

	case "quick":

	}
}
