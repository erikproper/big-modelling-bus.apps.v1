/*
 *
 * Module:      BIG Modelling Bus Apps, Version 1
 * Package:     Modelling Bus Apps
 * Application: Generic get application for the Modelling Bus, Version 1
 *
 * This is a generic application to get artefacts/observations/coordinations from the modelling bus.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 19.12.2025
 *
 */

package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

/*
 * Defining constants
 */

const (
	defaultIni = "config.ini" // Default configuration file name

	rawArtefactRetrieval         = "raw_artefact"         // Raw artefact retrieval kind
	jsonArtefactRetrieval        = "json_artefact"        // JSON artefact retrieval kind
	rawObservationRetrieval      = "raw_observation"      // Raw observation retrieval kind
	jsonObservationRetrieval     = "json_observation"     // JSON observation retrieval kind
	streamedObservationRetrieval = "streamed_observation" // Streamed observation retrieval kind
	coordinationRetrieval        = "coordination"         // Coordination retrieval kind

	timestampExtension = ".timestamp"
)

/*
 * Key variables
 */

var (
	modellingBusConnector connect.TModellingBusConnector // The Modelling Bus Connector

	localFilePath string // The local file path to store retrieved artefact

	// Handlers for different retrieval kinds
	retrievalHandlers = map[string]func(){
		rawArtefactRetrieval:         handleRawArtefactRetrieval,         // Handler for raw artefact retrieval
		jsonArtefactRetrieval:        handleJSONArtefactRetrieval,        // Handler for JSON artefact retrieval
		rawObservationRetrieval:      handleRawObservationRetrieval,      // Handler for raw observation retrieval
		jsonObservationRetrieval:     handleJSONObservationRetrieval,     // Handler for JSON observation retrieval
		streamedObservationRetrieval: handleStreamedObservationRetrieval, // Handler for streamed observation retrieval
		coordinationRetrieval:        handleCoordinationRetrieval,        // Handler for coordination retrieval
	}

	// Explaining the retrieval kind flag
	retrievalKindExplain = "Kind of retrieval to conduct. One of: " +
		rawArtefactRetrieval + ", " +
		jsonArtefactRetrieval + ", " +
		rawObservationRetrieval + ", " +
		jsonObservationRetrieval + ", " +
		streamedObservationRetrieval + ", or " +
		coordinationRetrieval + "."

	configFlag            = flag.String("config", defaultIni, "Configuration file")                  // Configuration file flag
	reportLevelFlag       = flag.Int("reporting", generics.ProgressLevelBasic, "Reporting level")    // Reporting level flag
	agentIDFlag           = flag.String("agent_id", "", "Agent ID")                                  // Agent ID flag
	fileNameFlag          = flag.String("file_name", "", "Local file name to store retrieved files") // Local file name flag
	observationIDFlag     = flag.String("observation_id", "", "Observation ID")                      // Observation ID flag
	coordinationTopicFlag = flag.String("coordination_topic", "", "Coordination topic path")         // Coordination topic path flag
	retrievalKindFlag     = flag.String("kind", "", retrievalKindExplain)                            // Retrieval kind flag
	jsonVersionFlag       = flag.String("json_version", "", "JSON version of JSON artefact content") // JSON version flag
	artefactIDFlag        = flag.String("artefact_id", "", "Artefact ID")                            // Artefact ID flag
	waitFlag              = flag.Bool("wait", false, "wait for a posting")                           // Wait flag
	waitModeFlag          = flag.String("wait_mode", "", "wait mode when waiting for a posting")     // Wait mode flag
)

/*
 * Generic functionality to support the retrieval handlers
 */

// Write timestamp to a file
func writeTimestampToFile(timestamp, filePath string) {
	if err := os.WriteFile(filePath+timestampExtension, []byte(timestamp), 0644); err != nil {
		// Reporting error
		modellingBusConnector.Reporter.ReportError("Error writing to timestamp file:", err)
	}
}

// Save JSON to file with given kind and base file name
func SaveJSONToFile(jsonContent []byte, timestamp, kind string) {
	fileBaseName := *fileNameFlag + generics.JSONExtension

	if len(kind) > 0 {
		fileBaseName = kind + "_" + fileBaseName
	}

	filePath := filepath.FromSlash(localFilePath + "/" + fileBaseName)
	if err := os.WriteFile(filePath, jsonContent, 0644); err != nil {
		// Reporting error
		modellingBusConnector.Reporter.ReportError("Error writing to json file:", err)
		return
	}

	// Write timestamp to a file
	writeTimestampToFile(timestamp, filePath)

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Retrieved JSON artefact for %s as: %s", kind, filePath)
}

// Deferred or immediate retrieval
func deferredOrImmediate(progress string, deferredHandler func(*bool), immediateHandler func()) {
	if *waitFlag {
		// Reporting progress
		modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Deferred %s retrieval.", progress)

		finished := false

		deferredHandler(&finished)

		for !finished {
			time.Sleep(1 * time.Second)
		}
	} else {
		// Reporting progress
		modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Immediate %s retrieval.", progress)

		immediateHandler()
	}
}

/*
 * Handlers for different retrieval kinds
 */

// Handler for raw artefact retrieval
func handleRawArtefactRetrieval() {
	// We need an artefact ID for artefact retrievals
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for artefact retrieval.") {
		return
	}

	// We must also have an agent ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(agentIDFlag, "No agent ID specified.") {
		return
	}

	// Create the modelling bus artefact retriever
	modellingBusArtefactRetriever := connect.CreateModellingBusArtefactConnector(modellingBusConnector, "", *artefactIDFlag)

	// Deferred or immediate variation
	deferredOrImmediate("raw artefact",
		func(finished *bool) {
			// Deferr for a raw artefact state posting
			modellingBusArtefactRetriever.ListenForRawArtefactStatePostings(*agentIDFlag, *artefactIDFlag, *fileNameFlag, func(filePath string, timestamp string) {
				// Write timestamp to a file
				writeTimestampToFile(timestamp, filePath)

				// Reporting progress
				modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Retrieved raw artefact as: %s", filePath)

				*finished = true
			})
		},
		func() {
			// Retrieving the raw artefact
			filePath, timestamp := modellingBusArtefactRetriever.GetRawArtefact(*agentIDFlag, *artefactIDFlag, *fileNameFlag)

			// Write timestamp to a file
			writeTimestampToFile(timestamp, filePath)

			// Reporting progress
			modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Retrieved raw artefact as: %s", filePath)
		})
}

// Handler for JSON artefact retrieval
func handleJSONArtefactRetrieval() {
	// We need a JSON version for JSON artefact retrievals
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(jsonVersionFlag, "No JSON version specified for JSON artefact retrieval.") {
		return
	}

	// We must also have an agent ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(agentIDFlag, "No agent ID specified.") {
		return
	}

	// We also need an artefact ID for artefact retrievals
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(artefactIDFlag, "No artefact ID specified for artefact retrieval.") {
		return
	}

	// Create the modelling bus artefact retriever
	modellingBusArtefactRetriever := connect.CreateModellingBusArtefactConnector(modellingBusConnector, *jsonVersionFlag, *artefactIDFlag)

	deferredOrImmediate("JSON artefact",
		func(finished *bool) {
			if *waitModeFlag == "state" {
				modellingBusArtefactRetriever.ListenForJSONArtefactStatePostings(*agentIDFlag, *artefactIDFlag, func() {
					SaveJSONToFile(modellingBusArtefactRetriever.CurrentContent, modellingBusArtefactRetriever.CurrentTimestamp, "state")
					*finished = true
				})
				modellingBusArtefactRetriever.ListenForJSONArtefactUpdatePostings(*agentIDFlag, *artefactIDFlag, func() {})
				modellingBusArtefactRetriever.ListenForJSONArtefactConsideringPostings(*agentIDFlag, *artefactIDFlag, func() {})

			} else if *waitModeFlag == "update" {
				modellingBusArtefactRetriever.ListenForJSONArtefactStatePostings(*agentIDFlag, *artefactIDFlag, func() {})
				modellingBusArtefactRetriever.ListenForJSONArtefactUpdatePostings(*agentIDFlag, *artefactIDFlag, func() {
					SaveJSONToFile(modellingBusArtefactRetriever.UpdatedContent, modellingBusArtefactRetriever.UpdatedTimestamp, "update")
					*finished = true
				})
				modellingBusArtefactRetriever.ListenForJSONArtefactConsideringPostings(*agentIDFlag, *artefactIDFlag, func() {})

			} else if *waitModeFlag == "considering" {
				modellingBusArtefactRetriever.ListenForJSONArtefactStatePostings(*agentIDFlag, *artefactIDFlag, func() {})
				modellingBusArtefactRetriever.ListenForJSONArtefactUpdatePostings(*agentIDFlag, *artefactIDFlag, func() {})
				modellingBusArtefactRetriever.ListenForJSONArtefactConsideringPostings(*agentIDFlag, *artefactIDFlag, func() {
					SaveJSONToFile(modellingBusArtefactRetriever.ConsideredContent, modellingBusArtefactRetriever.ConsideredTimestamp, "considered")
					*finished = true
				})

			} else {
				modellingBusArtefactRetriever.ListenForJSONArtefactStatePostings(*agentIDFlag, *artefactIDFlag, func() {
					SaveJSONToFile(modellingBusArtefactRetriever.CurrentContent, modellingBusArtefactRetriever.CurrentTimestamp, "state")
					SaveJSONToFile(modellingBusArtefactRetriever.UpdatedContent, modellingBusArtefactRetriever.UpdatedTimestamp, "update")
					SaveJSONToFile(modellingBusArtefactRetriever.ConsideredContent, modellingBusArtefactRetriever.ConsideredTimestamp, "considered")
					*finished = true
				})
				modellingBusArtefactRetriever.ListenForJSONArtefactUpdatePostings(*agentIDFlag, *artefactIDFlag, func() {
					SaveJSONToFile(modellingBusArtefactRetriever.CurrentContent, modellingBusArtefactRetriever.CurrentTimestamp, "state")
					SaveJSONToFile(modellingBusArtefactRetriever.UpdatedContent, modellingBusArtefactRetriever.UpdatedTimestamp, "update")
					SaveJSONToFile(modellingBusArtefactRetriever.ConsideredContent, modellingBusArtefactRetriever.ConsideredTimestamp, "considered")
					*finished = true
				})
				modellingBusArtefactRetriever.ListenForJSONArtefactConsideringPostings(*agentIDFlag, *artefactIDFlag, func() {
					SaveJSONToFile(modellingBusArtefactRetriever.CurrentContent, modellingBusArtefactRetriever.CurrentTimestamp, "state")
					SaveJSONToFile(modellingBusArtefactRetriever.UpdatedContent, modellingBusArtefactRetriever.UpdatedTimestamp, "update")
					SaveJSONToFile(modellingBusArtefactRetriever.ConsideredContent, modellingBusArtefactRetriever.ConsideredTimestamp, "considered")
					*finished = true
				})
			}
		},
		func() {
			// Retrieving the JSON artefact state, update, and considering
			modellingBusArtefactRetriever.GetJSONArtefactState(*agentIDFlag, *artefactIDFlag)
			modellingBusArtefactRetriever.GetJSONArtefactUpdate(*agentIDFlag, *artefactIDFlag)
			modellingBusArtefactRetriever.GetJSONArtefactConsidering(*agentIDFlag, *artefactIDFlag)

			// Save JSONs to files
			SaveJSONToFile(modellingBusArtefactRetriever.CurrentContent, modellingBusArtefactRetriever.CurrentTimestamp, "state")
			SaveJSONToFile(modellingBusArtefactRetriever.UpdatedContent, modellingBusArtefactRetriever.UpdatedTimestamp, "update")
			SaveJSONToFile(modellingBusArtefactRetriever.ConsideredContent, modellingBusArtefactRetriever.ConsideredTimestamp, "considered")
		})
}

// Handler for raw observation retrieval
func handleRawObservationRetrieval() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// We must also have an agent ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(agentIDFlag, "No agent ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw observation retrieval.")

	// Retrieving the raw observation
	filePath, timestamp := modellingBusConnector.GetRawObservation(*agentIDFlag, *observationIDFlag, *fileNameFlag)

	// timestampFileNameFlag
	writeTimestampToFile(timestamp, filePath)

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Retrieved raw observation as: %s", filePath)
}

// Handler for JSON observation retrieval
func handleJSONObservationRetrieval() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// We must also have an agent ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(agentIDFlag, "No agent ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON observation retrieval.")

	// Retrieving the JSON observation
	observation, timestamp := modellingBusConnector.GetJSONObservation(*agentIDFlag, *observationIDFlag)

	// Saving the JSON observation to a file
	SaveJSONToFile(observation, timestamp, "")
}

// Handler for streamed observation retrieval
func handleStreamedObservationRetrieval() {
	// We must have an observation ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(observationIDFlag, "No observation ID specified.") {
		return
	}

	// We must also have an agent ID
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(agentIDFlag, "No agent ID specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Streamed observation retrieval.")

	// Retrieving the JSON observation
	observation, timestamp := modellingBusConnector.GetStreamedObservation(*agentIDFlag, *observationIDFlag)

	// Saving the JSON observation to a file
	SaveJSONToFile(observation, timestamp, "")
}

// Handler for coordination retrieval
func handleCoordinationRetrieval() {
	// We must have a coordination topic
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(coordinationTopicFlag, "No coordination topic specified.") {
		return
	}

	// Reporting progress
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination retrieval.")

	coordination, timestamp := modellingBusConnector.GetCoordination(*coordinationTopicFlag)

	// Saving the JSON observation to a file
	SaveJSONToFile(coordination, timestamp, "")
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

	// Getting the work folder
	localFilePath = configData.GetValue("", "work_folder").String()

	// Creating the Modelling Bus Connector
	modellingBusConnector = connect.CreateModellingBusConnector(configData, reporter, !connect.PostingOnly)

	// We must always have a retrieval kind
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(retrievalKindFlag, "No retrieval kind specified.") {
		return
	}

	// We also also, always have a file name
	if modellingBusConnector.Reporter.MaybeReportEmptyFlagError(fileNameFlag, "No file name specified for artefact retrieval.") {
		return
	}

	// Getting the retrieval handler
	retrievalHandler := retrievalHandlers[*retrievalKindFlag]

	// Validating retrieval handler
	if retrievalHandler == nil {
		modellingBusConnector.Reporter.Error("Unknown retrieval kind specified: %s.", *retrievalKindFlag)

		return
	}

	// Calling the retrieval handler
	retrievalHandler()
}
