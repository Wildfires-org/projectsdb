package main

import (
	"encoding/csv"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

type PALSToCsvConfig struct {
	PalsDataFile string `required help:"PALS Data CSV File" type:"path"`
}

func PalsToCsv(config PALSToCsvConfig) error {
	filePath := "data/pals.csv"
	err := writePALSCsv(config.PalsDataFile, filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filePath,
			"error": err.Error(),
		}).Error("Unable to write CSV")
		return err
	}
	return nil

}

type PalsEntry struct {
	ProjectNumber                     string
	ForestId                          string
	ProjectName                       string
	LmuActual                         string
	LmuRegion                         string
	LmuForest                         string
	LmuDistrict                       string
	ProjectStatus                     string
	ProjectCreated                    string
	CreatedFy                         string
	DecisionId                        string
	DecisionName                      string
	InitiationDate                    string
	InitiationFy                      string
	DecisionSigned                    string
	SignedFy                          string
	SignerLastName                    string
	SignerFirstName                   string
	SignerTitle                       string
	DecisionType                      string
	DecisionAppealRule                string
	ProjectNoticeAndCommentRegulation string
	AppealedOrObjected                string
	NoCommentsOrOnlySupport           string
	Litigated                         string
	FactsActivity                     string
	Activities                        string
	Purposes                          string
	UniqueProject                     string
	ElapsedDays                       string
	UniqueDecision                    string
	Ongoing                           string
	DistrictId                        string
	RegionId                          string
	DecisionLevel                     string
	RegionName                        string
	Forest                            string
	CalendarYearSigned                string
	CalendarYearInitiated             string
	OverallCaseOutcome                string
	CaseStatus                        string
}

func (palsRow PalsEntry) AsCsv() []string {
	return []string{
		palsRow.ProjectNumber,
		palsRow.ForestId,
		palsRow.ProjectName,
		palsRow.LmuActual,
		palsRow.LmuRegion,
		palsRow.LmuForest,
		palsRow.LmuDistrict,
		palsRow.ProjectStatus,
		palsRow.ProjectCreated,
		palsRow.CreatedFy,
		palsRow.DecisionId,
		palsRow.DecisionName,
		palsRow.InitiationDate,
		palsRow.InitiationFy,
		palsRow.DecisionSigned,
		palsRow.SignedFy,
		palsRow.SignerLastName,
		palsRow.SignerFirstName,
		palsRow.SignerTitle,
		palsRow.DecisionType,
		palsRow.DecisionAppealRule,
		palsRow.ProjectNoticeAndCommentRegulation,
		palsRow.AppealedOrObjected,
		palsRow.NoCommentsOrOnlySupport,
		palsRow.Litigated,
		palsRow.FactsActivity,
		palsRow.Activities,
		palsRow.Purposes,
		palsRow.UniqueProject,
		palsRow.ElapsedDays,
		palsRow.UniqueDecision,
		palsRow.Ongoing,
		palsRow.DistrictId,
		palsRow.RegionId,
		palsRow.DecisionLevel,
		palsRow.RegionName,
		palsRow.Forest,
		palsRow.CalendarYearSigned,
		palsRow.CalendarYearInitiated,
		palsRow.OverallCaseOutcome,
		palsRow.CaseStatus,
	}
}

func writePALSCsv(original_path string, path string) error {
	// 26 - 43
	purposes := map[string]string{
		"FC": "Facility management",
		"FR": "Research",
		"HF": "Fuels management",
		"HR": "Heritage resource management",
		"LM": "Land ownership management",
		"LW": "Land acquisition",
		"MG": "Minerals and geology",
		"PN": "Land management planning",
		"RD": "Road management",
		"RG": "Grazing management",
		"RO": "Regulations, directives, orders",
		"RU": "Special area management",
		"RW": "Recreation management",
		"SU": "Special use management",
		"TM": "Forest products",
		"VM": "Vegetation management (non-forest products)",
		"WF": "Wildlife, fish, rare plants",
		"WM": "Water management",
	}
	// 44 - 52 and 54 - 93
	activities := map[string]string{
		"AL": "Land use adjustments",
		"BL": "Boundary adjustments",
		"BM": "Biomass",
		"CP": "Plan creation/revision",
		"DC": "Directive creation/modification",
		"DR": "Road decommissioning",
		"DS": "Developed site management",
		"EC": "Environmental compliance actions",
		"ET": "Electric transmission",
		"FI": "Facility improvements/construction",
		"FN": "Fuel treatments",
		"FV": "Forest vegetation improvements",
		"GA": "Dispersed recreation management",
		"GP": "Grazing allotment management",
		"GR": "Grazing authorizations",
		"GT": "Geothermal",
		"HI": "Species habitat improvements",
		"HP": "Hydropower",
		"HR": "Heritage resource management",
		"LA": "Special use authorizations",
		"LP": "Land purchases",
		"MF": "Facility maintenance",
		"ML": "Abandoned mine land clean-up",
		"MO": "Minerals or geology plans of operations",
		"MP": "Plan amendment",
		"MT": "Trail management",
		"NC": "Special products sales",
		"NG": "Natural gas",
		"NW": "Noxious weed treatments",
		"OC": "Order creation/modification",
		"OL": "Oil",
		"PE": "Species population enhancements",
		"PJ": "Land exchanges",
		"RA": "Roadless area management",
		"RC": "Regulation creation/modification",
		"RD": "Road maintenance",
		"RE": "Research and development",
		"RI": "Road improvements/construction",
		"RV": "Rangeland vegetation improvements",
		"SA": "Special area management",
		"SC": "Scenery management",
		"SI": "Grazing structural improvements",
		"SL": "Solar",
		"SS": "Timber sales (salvage)",
		"TR": "Travel management",
		"TS": "Timber sales (green)",
		"WC": "Watershed improvements",
		"WD": "Wilderness management",
		"WI": "Wind",
	}
	//open original file
	f, err := os.Open(original_path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var newPals []PalsEntry
	headers := data[0]
	for _, line := range data[1:] {
		// only get projects after 2012
		if yearAsInt, _ := strconv.Atoi(line[10]); yearAsInt < 2012 {
			continue
		}
		var currentActivities []string
		var currentPurposes []string
		// purposes
		for i, purposeBool := range line[26:43] {
			if purposeBool == "1" {
				purposeCode := headers[i+26]                                     // FC Facility Management
				purposeCode = purposeCode[0:2]                                   // FC
				currentPurposes = append(currentPurposes, purposes[purposeCode]) // Facility Management
			}
		}
		// activities
		for i, activityBool := range line[44:93] {
			if activityBool == "1" && i+44 != 53 { // need to avoid 53 as it's FACTS activity not a real activity
				activityCode := headers[i+44]                                           // BM Biomass
				activityCode = activityCode[0:2]                                        // BM
				currentActivities = append(currentActivities, activities[activityCode]) // Biomass
			}
		}
		newPals = append(newPals, PalsEntry{
			ProjectNumber:                     line[1],
			ForestId:                          line[2],
			ProjectName:                       line[3],
			LmuActual:                         line[4],
			LmuRegion:                         line[5],
			LmuForest:                         line[6],
			LmuDistrict:                       line[7],
			ProjectStatus:                     line[8],
			ProjectCreated:                    line[9],
			CreatedFy:                         line[10],
			DecisionId:                        line[11],
			DecisionName:                      line[12],
			InitiationDate:                    line[13],
			InitiationFy:                      line[14],
			DecisionSigned:                    line[15],
			SignedFy:                          line[16],
			SignerLastName:                    line[17],
			SignerFirstName:                   line[18],
			SignerTitle:                       line[19],
			DecisionType:                      line[20],
			DecisionAppealRule:                line[21],
			ProjectNoticeAndCommentRegulation: line[22],
			AppealedOrObjected:                line[23],
			NoCommentsOrOnlySupport:           line[24],
			Litigated:                         line[25],
			FactsActivity:                     line[53],
			Activities:                        strings.Join(currentActivities, ", "),
			Purposes:                          strings.Join(currentPurposes, ", "),
			UniqueProject:                     line[94],
			ElapsedDays:                       line[95],
			UniqueDecision:                    line[96],
			Ongoing:                           line[97],
			DistrictId:                        line[98],
			RegionId:                          line[99],
			DecisionLevel:                     line[100],
			RegionName:                        line[101],
			Forest:                            line[102],
			CalendarYearSigned:                line[103],
			CalendarYearInitiated:             line[104],
			OverallCaseOutcome:                line[105],
			CaseStatus:                        line[106],
		})
	}

	// open file
	csvFile, err := os.Create(path)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
		return err
	}
	defer csvFile.Close()

	// build column headers
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{
		"Project Number",
		"Forest Id",
		"Project Name",
		"Lmu Actual",
		"Lmu Region",
		"Lmu Forest",
		"Lmu District",
		"Project Status",
		"Project Created",
		"Created Fy",
		"Decision Id",
		"Decision Name",
		"Initiation Date",
		"Initiation Fy",
		"Decision Signed",
		"Signed Fy",
		"Signer Last Name",
		"Signer First Name",
		"Signer Title",
		"Decision Type",
		"Decision Appeal Rule",
		"Project Notice and Comment Regulation",
		"Appealed or Objected",
		"No Comments or Only Support",
		"Litigated",
		"FACTS Activity",
		"Activities",
		"Purposes",
		"Unique Project",
		"Elapsed Days",
		"Unique Decision",
		"Ongoing",
		"District Id",
		"Region Id",
		"Decision Level",
		"Region Name",
		"Forest",
		"Calendar Year Signed",
		"Calendar Year Initiated",
		"Overall Case Outcome",
		"Case Status",
	})

	//write rows to file
	for _, record := range newPals {
		if err := writer.Write(record.AsCsv()); err != nil {
			log.Fatalf("failed writing to csv: %s", err)
			return err
		}
	}
	writer.Flush()
	return nil
}
