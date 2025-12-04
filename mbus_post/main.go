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

	rawArtefactPosting  = "raw_artefact"
	jsonArtefactPosting = "json_artefact"
	observationPosting  = "observation"
	coordinationPosting = "coordination"
)

var (
	modellingBusConnector connect.TModellingBusConnector

	postingHandlers = map[string]func(){
		rawArtefactPosting:  handleRawArtefactPosting,
		jsonArtefactPosting: handleJSONArtefactPosting,
		observationPosting:  handleObservationPosting,
		coordinationPosting: handleCoordinationPosting,
	}

	postingKindExplain = "Kind of posting to make. One of: " +
		rawArtefactPosting + ", " +
		jsonArtefactPosting + ", " +
		observationPosting + ", or " +
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

func handleRawArtefactPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw artefact posting")

	if len(*fileFlag) == 0 {
		modellingBusConnector.Reporter.Error("No file specified for raw artefact posting.")

		return
	} else {
		modellingBusConnector.PostRawArtefact(*topicFlag, *fileFlag)
	}
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

	modellingBusJSONPoster := connect.CreateModellingBusJSONArtefactConnector(modellingBusConnector, *jsonVersionFlag)

	modellingBusJSONPoster.PrepareForPosting(*artefactIDFlag)

	if len(*jsonFlag) > 0 {
		modellingBusJSONPoster.PostState([]byte(*jsonFlag), nil)
	} else if len(*fileFlag) > 0 {
		jsonPayload, err := os.ReadFile(*fileFlag)
		if err != nil {
			modellingBusConnector.Reporter.Error("Error reading file for JSON artefact posting. %s", err)

			return
		} else {
			modellingBusJSONPoster.PostState(jsonPayload, nil)
		}
	}
}

func handleObservationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Observation posting")
}

func handleCoordinationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination posting")
}

func ReportProgress(message string) {
	fmt.Println("PROGRESS:", message)
}

func ReportError(message string) {
	fmt.Println("ERROR:", message)
}

func main() {
	flag.Parse()

	reporter := generics.CreateReporter(*reportLevelFlag, ReportError, ReportProgress)
	configData := generics.LoadConfig(*configFlag, reporter)
	modellingBusConnector = connect.CreateModellingBusConnector(configData, reporter)

	if len(*topicFlag) == 0 {
		reporter.Error("No topic path specified.")

		return
	}

	if len(*postingKindFlag) == 0 {
		reporter.Error("No posting kind specified.")

		return
	}

	if len(*fileFlag) == 0 && len(*jsonFlag) == 0 {
		modellingBusConnector.Reporter.Error("Need file or JSON record.")

		return
	}

	postingHandler := postingHandlers[*postingKindFlag]

	if postingHandler == nil {
		reporter.Error("Unknown posting kind specified: %s.", *postingKindFlag)

		return
	}

	postingHandler()
}
