/*
 *
 * Module:      BIG Modelling Bus Apps, Version 1
 * Package:     Modelling Bus Apps
 * Application: XX
 *
 * XXXX
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 18.12.2025
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

/*
 * Defining constants
 */

const (
	defaultIni = "config.ini" // Default configuration file name

	rawArtefactPosting         = "raw_artefact"         // Raw artefact posting kinds
	jsonArtefactPosting        = "json_artefact"        // JSON artefact posting kinds
	rawObservationPosting      = "raw_observation"      // Raw observation posting kinds
	jsonObservationPosting     = "json_observation"     // JSON observation posting kinds
	streamedObservationPosting = "streamed_observation" // Streamed observation posting kinds
	coordinationPosting        = "coordination"         // Coordination posting kinds
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

	configFlag      = flag.String("config", defaultIni, "Configuration file")                   // Configuration file flag
	reportLevelFlag = flag.Int("reporting", generics.ProgressLevelBasic, "Reporting level")     // Reporting level flag
	topicFlag       = flag.String("topic", "", "Topic path")                                    // Topic path flag
	postingKindFlag = flag.String("kind", "", postingKindExplain)                               // Posting kind flag
	fileFlag        = flag.String("file", "", "File to post")                                   // File to post flag
	jsonFlag        = flag.String("json", "", "JSON content to post")                           // JSON content to post flag
	jsonVersionFlag = flag.String("json_version", "", "JSON version of JSON artefact content.") // JSON version flag
	artefactIDFlag  = flag.String("artefact_id", "", "Artefact ID of JSON artefact content.")   // Artefact ID flag
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

	// Create the modelling bus artefact poster
	modellingBusArtefactPoster := connect.CreateModellingBusArtefactConnector(modellingBusConnector, *jsonVersionFlag, *artefactIDFlag)

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw artefact posting.")

	// Posting the raw artefact
	modellingBusArtefactPoster.PostRawArtefactState(*topicFlag, *fileFlag)
}

// Handling JSON artefact posting
func handleJSONArtefactPosting() {
	// We need a JSON version for JSON artefact posting
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(jsonVersionFlag, "No JSON version specified for JSON artefact posting.") {
		return
	}

	// We also need an artefact ID for JSON artefact postings
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for JSON artefact posting.") {
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

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw observation posting.")

	// Posting the raw observation
	modellingBusConnector.PostRawObservation(*topicFlag, *fileFlag)
}

// Handling JSON observation posting
func handleJSONObservationPosting() {
	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON observation posting.")

	// Posting the JSON observation
	modellingBusConnector.PostJSONObservation(*topicFlag, jsonPayload)
}

// Handling streamed observation posting
func handleStreamedObservationPosting() {
	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Streamed observation posting.")

	// Posting the streamed observation
	modellingBusConnector.PostStreamedObservation(*topicFlag, jsonPayload)
}

func handleCoordinationPosting() {
	// Getting the JSON payload
	jsonPayload, ok := getJSONPayload()

	// Checking if we got the payload properly
	if !ok {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination posting.")

	// Posting the coordination
	modellingBusConnector.PostCoordination(*topicFlag, jsonPayload)
}

/*
 * Reporting progress and errors
 */

func ReportProgress(message string) {
	fmt.Println("PROGRESS:", message)
}

func ReportError(message string) {
	fmt.Println("ERROR:", message)
}

/*
 * Main function
 */

func main() {
	// Parsing flags
	flag.Parse()

	// Creating the reporter
	reporter := generics.CreateReporter(*reportLevelFlag, ReportError, ReportProgress)

	// Loading the configuration
	configData := generics.LoadConfig(*configFlag, reporter)

	// Creating the Modelling Bus Connector
	modellingBusConnector = connect.CreateModellingBusConnector(configData, reporter, connect.PostingOnly)

	// We must have a topic path
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(topicFlag, "No topic path specified.") {
		return
	}

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
