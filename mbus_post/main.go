package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

const (
	defaultIni = "config.ini"

	rawArtefactPosting         = "raw_artefact"
	jsonArtefactPosting        = "json_artefact"
	rawObservationPosting      = "raw_observation"
	jsonObservationPosting     = "json_observation"
	streamedObservationPosting = "streamed_observation"
	coordinationPosting        = "coordination"
)

var (
	modellingBusConnector connect.TModellingBusConnector

	postingHandlers = map[string]func(){
		rawArtefactPosting:         handleRawArtefactPosting,
		jsonArtefactPosting:        handleJSONArtefactPosting,
		rawObservationPosting:      handleRawObservationPosting,
		jsonObservationPosting:     handleJSONObservationPosting,
		streamedObservationPosting: handleStreamedObservationPosting,
		coordinationPosting:        handleCoordinationPosting,
	}

	postingKindExplain = "Kind of posting to make. One of: " +
		rawArtefactPosting + ", " +
		jsonArtefactPosting + ", " +
		rawObservationPosting + ", " +
		jsonObservationPosting + ", " +
		streamedObservationPosting + ", or " +
		coordinationPosting + "."

	configFlag      = flag.String("config", defaultIni, "Configuration file")
	reportLevelFlag = flag.Int("reporting", 1, "Reporting level")
	topicFlag       = flag.String("topic", "", "Topic path")
	postingKindFlag = flag.String("kind", "", postingKindExplain)
	fileFlag        = flag.String("file", "", "File to post")
	jsonFlag        = flag.String("json", "", "JSON content to post")
	jsonVersionFlag = flag.String("json_version", "", "JSON version of JSON artefact content.")
	artefactIDFlag  = flag.String("artefact_id", "", "Artefact ID of JSON artefact content.")
)

func ReportProgress(message string) {
	fmt.Println("PROGRESS:", message)
}

func ReportError(message string) {
	fmt.Println("ERROR:", message)
}

func getJSONPayload() ([]byte, bool) {
	jsonPayload := []byte(*jsonFlag)

	if len(*fileFlag) > 0 {
		err := error(nil)
		jsonPayload, err = os.ReadFile(*fileFlag)

		if err != nil {
			modellingBusConnector.Reporter.Error("Error reading file for JSON artefact posting. %s", err)

			return jsonPayload, false
		}
	}

	return jsonPayload, true
}

func handleRawArtefactPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw artefact posting")

	modellingBusArtefactPoster := connect.CreateModellingBusArtefactConnector(modellingBusConnector, *jsonVersionFlag)

	modellingBusArtefactPoster.PrepareForPosting(*artefactIDFlag)

	if len(*fileFlag) == 0 {
		modellingBusConnector.Reporter.Error("No file specified for raw artefact posting.")

		return
	}

	modellingBusArtefactPoster.PostRawArtefactState(*topicFlag, *fileFlag)
}

func handleJSONArtefactPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON artefact posting")

	if len(*jsonVersionFlag) == 0 {
		modellingBusConnector.Reporter.Error("No JSON version specified for JSON artefact posting.")

		return
	}

	if len(*artefactIDFlag) == 0 {
		modellingBusConnector.Reporter.Error("No artefact ID specified for JSON artefact posting.")

		return
	}

	modellingBusArtefactPoster := connect.CreateModellingBusArtefactConnector(modellingBusConnector, *jsonVersionFlag)

	modellingBusArtefactPoster.PrepareForPosting(*artefactIDFlag)

	if jsonPayload, ok := getJSONPayload(); ok {
		modellingBusArtefactPoster.PostJSONArtefactState(jsonPayload, nil)
	}
}

func handleRawObservationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw observation posting")

	if len(*fileFlag) == 0 {
		modellingBusConnector.Reporter.Error("No file specified for raw artefact posting.")

		return
	}

	modellingBusConnector.PostRawObservation(*topicFlag, *fileFlag)
}

func handleJSONObservationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON observation posting")

	if jsonPayload, ok := getJSONPayload(); ok {
		modellingBusConnector.PostJSONObservation(*topicFlag, jsonPayload)
	}
}

func handleStreamedObservationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON observation posting")

	if jsonPayload, ok := getJSONPayload(); ok {
		modellingBusConnector.PostStreamedObservation(*topicFlag, jsonPayload)
	}
}

func handleCoordinationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination posting")

	if jsonPayload, ok := getJSONPayload(); ok {
		modellingBusConnector.PostCoordination(*topicFlag, jsonPayload)
	}
}

func main() {
	flag.Parse()

	reporter := generics.CreateReporter(*reportLevelFlag, ReportError, ReportProgress)
	configData := generics.LoadConfig(*configFlag, reporter)
	modellingBusConnector = connect.CreateModellingBusConnector(configData, reporter, connect.PostingOnly)

	if len(*topicFlag) == 0 {
		reporter.Error("No topic path specified.")

		return
	}

	if len(*postingKindFlag) == 0 {
		reporter.Error("No posting kind specified.")

		return
	}

	postingHandler := postingHandlers[*postingKindFlag]

	if postingHandler == nil {
		reporter.Error("Unknown posting kind specified: %s.", *postingKindFlag)

		return
	}

	postingHandler()
}
