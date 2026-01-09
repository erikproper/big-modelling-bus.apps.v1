/*
 *
 * Module:      BIG Modelling Bus Apps, Version 1
 * Package:     Modelling Bus Apps
 * Application: Generic deleter for the Modelling Bus, Version 1
 *
 * This is a generic delete application for the modelling bus.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 18.12.2025
 *
 */

package main

import (
	"flag"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

/*
 * Defining constants
 */

const (
	defaultIni = "config.ini" // Default configuration file name

	rawArtefactDeletion         = "raw_artefact"         // Raw artefact deletion kind
	jsonArtefactDeletion        = "json_artefact"        // JSON artefact deletion kind
	rawObservationDeletion      = "raw_observation"      // Raw observation deletion kind
	jsonObservationDeletion     = "json_observation"     // JSON observation deletion kind
	streamedObservationDeletion = "streamed_observation" // Streamed observation deletion kind
	coordinationDeletion        = "coordination"         // Coordination deletion kind
	environmentDeletion         = "environment"          // Environment deletion kind
)

/*
 * Key variables
 */

var (
	modellingBusConnector connect.TModellingBusConnector // The Modelling Bus Connector

	// Handlers for different deletion kinds
	deletionHandlers = map[string]func(){
		rawArtefactDeletion:         handleRawArtefactDeletion,         // Handler for raw artefact deletion
		jsonArtefactDeletion:        handleJSONArtefactDeletion,        // Handler for JSON artefact deletion
		rawObservationDeletion:      handleRawObservationDeletion,      // Handler for raw observation deletion
		jsonObservationDeletion:     handleJSONObservationDeletion,     // Handler for JSON observation deletion
		streamedObservationDeletion: handleStreamedObservationDeletion, // Handler for streamed observation deletion
		coordinationDeletion:        handleCoordinationDeletion,        // Handler for coordination deletion
		environmentDeletion:         handleEnvironmentDeletion,         // Handler for environment deletion
	}

	// Explaining the deletion kind flag
	deletionKindExplain = "Kind of deletion to make. One of: " +
		rawArtefactDeletion + ", " +
		jsonArtefactDeletion + ", " +
		rawObservationDeletion + ", " +
		jsonObservationDeletion + ", " +
		streamedObservationDeletion + ", " +
		coordinationDeletion + ", or " +
		environmentDeletion + "."

	configFlag            = flag.String("config", defaultIni, "Configuration file")                  // Configuration file flag
	reportLevelFlag       = flag.Int("reporting", generics.ProgressLevelBasic, "Reporting level")    // Reporting level flag
	observationIDFlag     = flag.String("observation_id", "", "Observation ID")                      // Observation ID flag
	coordinationTopicFlag = flag.String("coordination_topic", "", "Coordination topic path")         // Coordination topic path flag
	deletionKindFlag      = flag.String("kind", "", deletionKindExplain)                             // Deletion kind flag
	jsonVersionFlag       = flag.String("json_version", "", "JSON version of JSON artefact content") // JSON version flag
	artefactIDFlag        = flag.String("artefact_id", "", "Artefact ID")                            // Artefact ID flag
	environmentFlag       = flag.String("environment", "", "Environment")                            // Environment flag
)

/*
 * Handlers for different deletion kinds
 */

// Handler for raw artefact deletion
func handleRawArtefactDeletion() {
	// We need an artefact ID for artefact deletions
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for artefact deletion.") {
		return
	}

	// Create the modelling bus artefact deleter
	modellingBusArtefactDeleter := connect.CreateModellingBusArtefactConnector(modellingBusConnector, "", *artefactIDFlag)

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw artefact deletion.")

	// Deleting the raw artefact
	modellingBusArtefactDeleter.DeleteRawArtefact(*artefactIDFlag)
}

// Handler for JSON artefact deletion
func handleJSONArtefactDeletion() {
	// We need a JSON version for JSON artefact deletions
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(jsonVersionFlag, "No JSON version specified for JSON artefact deletion.") {
		return
	}

	// We also need an artefact ID for artefact deletions
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for artefact deletion.") {
		return
	}

	// Create the modelling bus artefact deleter
	modellingBusArtefactDeleter := connect.CreateModellingBusArtefactConnector(modellingBusConnector, *jsonVersionFlag, *artefactIDFlag)

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON artefact deletion.")

	// Deleting the JSON artefact
	modellingBusArtefactDeleter.DeleteJSONArtefact(*artefactIDFlag)
}

// Handler for raw observation deletion
func handleRawObservationDeletion() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw observation deletion.")

	// Posting the raw observation
	modellingBusConnector.DeleteRawObservation(*observationIDFlag)
}

// Handler for JSON observation deletion
func handleJSONObservationDeletion() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON observation deletion.")

	// Deleting the JSON observation
	modellingBusConnector.DeleteJSONObservation(*observationIDFlag)
}

// Handler for streamed observation deletion
func handleStreamedObservationDeletion() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Streamed observation deletion.")

	// Deleting the streamed observation
	modellingBusConnector.DeleteStreamedObservation(*observationIDFlag)
}

// Handler for coordination deletion
func handleCoordinationDeletion() {
	// We must have a coordination topic
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(coordinationTopicFlag, "No coordination topic specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination deletion.")

	// Deleting the coordination
	modellingBusConnector.DeleteCoordination(*coordinationTopicFlag)

}

// Handler for environment deletion
func handleEnvironmentDeletion() {
	// We must have an environment flag
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(environmentFlag, "No environment specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Environment deletion.")

	// Deleting the environment
	modellingBusConnector.DeleteEnvironment(*environmentFlag)
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
	modellingBusConnector = connect.CreateModellingBusConnector(configData, reporter, !connect.PostingOnly)

	// We must have a deletion kind
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(deletionKindFlag, "No deletion kind specified.") {
		return
	}

	// Getting the deletion handler
	deletionHandler := deletionHandlers[*deletionKindFlag]

	// Validating deletion handler
	if deletionHandler == nil {
		modellingBusConnector.Reporter.Error("Unknown deletion kind specified: %s.", *deletionKindFlag)

		return
	}

	// Calling the deletion handler
	deletionHandler()
}
