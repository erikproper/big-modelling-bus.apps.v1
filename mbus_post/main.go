package main

import (
	"flag"
	"fmt"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

var (
	modellingBusConnector connect.TModellingBusConnector
)

func handleRawArtefactPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Raw artefact posting")
}

func handleJSONArtefactPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "JSON artefact posting")
}

func handleObservationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Observation posting")
}

func handleCoordinationPosting() {
	modellingBusConnector.Reporter.Progress(generics.ProgressLevelBasic, "Coordination posting")
}

const (
	defaultIni = "config.ini"

	rawArtefactPosting  = "raw_artefact"
	jsonArtefactPosting = "json_artefact"
	observationPosting  = "observation"
	coordinationPosting = "coordination"
)

var (
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
)

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

	if postingHandler, postingKindExists := postingHandlers[*postingKindFlag]; !postingKindExists {
		reporter.Error("Unknown posting kind specified: %s. %s", *postingKindFlag, postingKindExplain)

		return
	} else {
		postingHandler()
	}

	fmt.Println("T", *topicFlag)
	fmt.Println("K", *postingKindFlag)
	fmt.Println("F", *fileFlag)
	fmt.Println("J", *jsonFlag)
	fmt.Println(modellingBusConnector)
}
