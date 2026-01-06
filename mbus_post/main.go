/*
 *
 * Module:      BIG Modelling Bus Apps, Version 1
 * Package:     Modelling Bus Apps
 * Application: Generic Poster for the Modelling Bus, Version 1
 *
 * This is a generic poster application for the modelling bus.
 * It can post different kinds of artefacts, observations, and coordination messages.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 18.12.2025
 *
 */

package main

import (
	"flag"
	"os"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

/*
 * Defining constants
 */

const (
	defaultIni = "config.ini" // Default configuration file name

	rawArtefactPosting         = "raw_artefact"         // Raw artefact posting kind
	jsonArtefactPosting        = "json_artefact"        // JSON artefact posting kind
	rawObservationPosting      = "raw_observation"      // Raw observation posting kind
	jsonObservationPosting     = "json_observation"     // JSON observation posting kind
	streamedObservationPosting = "streamed_observation" // Streamed observation posting kinds
	coordinationPosting        = "coordination"         // Coordination posting kind
)

/*
 * Key variables
 */

var (
	modellingBusConnector connect.TModellingBusConnector // The Modelling Bus Connector

	// Handlers for different posting kinds
	postingHandlers = map[string]func(){
		rawArtefactPosting:         handleRawArtefactPosting,         // Handler for raw artefact posting
		jsonArtefactPosting:        handleJSONArtefactPosting,        // Handler for JSON artefact posting
		rawObservationPosting:      handleRawObservationPosting,      // Handler for raw observation posting
		jsonObservationPosting:     handleJSONObservationPosting,     // Handler for JSON observation posting
		streamedObservationPosting: handleStreamedObservationPosting, // Handler for streamed observation posting
		coordinationPosting:        handleCoordinationPosting,        // Handler for coordination posting
	}

	// Explaining the posting kind flag
	postingKindExplain = "Kind of posting to make. One of: " +
		rawArtefactPosting + ", " +
		jsonArtefactPosting + ", " +
		rawObservationPosting + ", " +
		jsonObservationPosting + ", " +
		streamedObservationPosting + ", or " +
		coordinationPosting + "."

	configFlag            = flag.String("config", defaultIni, "Configuration file")                  // Configuration file flag
	reportLevelFlag       = flag.Int("reporting", generics.ProgressLevelBasic, "Reporting level")    // Reporting level flag
	observationIDFlag     = flag.String("observation_id", "", "Observation ID")                      // Observation ID flag
	coordinationTopicFlag = flag.String("coordination_topic", "", "Coordination topic path")         // Coordination topic path flag
	postingKindFlag       = flag.String("kind", "", postingKindExplain)                              // Posting kind flag
	fileFlag              = flag.String("file", "", "File to post")                                  // File to post flag
	jsonFlag              = flag.String("json", "", "JSON content to post")                          // JSON content to post flag
	jsonVersionFlag       = flag.String("json_version", "", "JSON version of JSON artefact content") // JSON version flag
	artefactIDFlag        = flag.String("artefact_id", "", "Artefact ID")                            // Artefact ID flag
)

/*
 * Getting the JSON payload to post
 */

func getJSONPayload() ([]byte, bool) {
	// Getting the JSON payload
	jsonPayload := []byte(*jsonFlag)

	// If no JSON content is given, we try to read it from a file
	if len(jsonPayload) == 0 && len(*fileFlag) > 0 {
		var err error

		// Reading the file content
		jsonPayload, err = os.ReadFile(*fileFlag)

		// Reporting errors if needed
		if modellingBusConnector.Reporter.MaybeReportError("Error reading file for JSON artefact posting:", err) {
			return []byte{}, false
		}
	}

	return jsonPayload, true
}

/*
 * Handlers for different posting kinds
 */

// Handling raw artefact posting
func handleRawArtefactPosting() {
	// Check if we have a file to post
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(fileFlag, "No file specified for raw artefact posting.") {
		return
	}

	// We also need an artefact ID for artefact postings
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for artefact posting.") {
		return
	}

	// Create the modelling bus artefact poster
	modellingBusArtefactPoster := connect.CreateModellingBusArtefactConnector(modellingBusConnector, "", *artefactIDFlag)

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw artefact posting.")

	// Posting the raw artefact
	modellingBusArtefactPoster.PostRawArtefactState(*fileFlag)
}

// Handling JSON artefact posting
func handleJSONArtefactPosting() {
	// We need a JSON version for JSON artefact posting
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(jsonVersionFlag, "No JSON version specified for JSON artefact posting.") {
		return
	}

	// We also need an artefact ID for artefact postings
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for artefact posting.") {
		return
	}

	// Creating modelling bus artefact poster
	modellingBusArtefactPoster := connect.CreateModellingBusArtefactConnector(modellingBusConnector, *jsonVersionFlag, *artefactIDFlag)

	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON artefact posting.")

	// Posting the JSON artefact
	modellingBusArtefactPoster.PostJSONArtefactState(jsonPayload, ok)
}

// Handling raw observation posting
func handleRawObservationPosting() {
	// Check if we have a file to post
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(fileFlag, "No file specified for raw observation posting.") {
		return
	}

	// We must have a topic path
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw observation posting.")

	// Posting the raw observation
	modellingBusConnector.PostRawObservation(*observationIDFlag, *fileFlag)
}

// Handling JSON observation posting
func handleJSONObservationPosting() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON observation posting.")

	// Posting the JSON observation
	modellingBusConnector.PostJSONObservation(*observationIDFlag, jsonPayload)
}

// Handling streamed observation posting
func handleStreamedObservationPosting() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Streamed observation posting.")

	// Posting the streamed observation
	modellingBusConnector.PostStreamedObservation(*observationIDFlag, jsonPayload)
}

func handleCoordinationPosting() {
	// We must have a coordination topic
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(coordinationTopicFlag, "No coordination topic specified.") {
		return
	}

	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination posting.")

	// Posting the coordination
	modellingBusConnector.PostCoordination(*coordinationTopicFlag, jsonPayload)
}

/*
 * Main function
 */

func main() {
	// Parsing flags
	flag.Parse()

	// Creating the reporter
	reporter := generics.CreateReporter(*reportLevelFlag, generics.ReportError, generics.ReportProgress)

	// Loading the configuration
	configData := generics.LoadConfig(*configFlag, reporter)

	// Creating the Modelling Bus Connector
	modellingBusConnector = connect.CreateModellingBusConnector(configData, reporter, connect.PostingOnly)

	// We must have a posting kind
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(postingKindFlag, "No posting kind specified.") {
		return
	}

	// Getting the posting handler
	postingHandler := postingHandlers[*postingKindFlag]

	// Validating posting handler
	if postingHandler == nil {
		modellingBusConnector.Reporter.Error("Unknown posting kind specified: %s.", *postingKindFlag)

		return
	}

	// Calling the posting handler
	postingHandler()
}
