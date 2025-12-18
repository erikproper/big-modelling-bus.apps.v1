/*
 *
 * Module:      BIG Modelling Bus Apps, Version 1
 * Package:     Modelling Bus Apps
 * Application: Poster for CDM Models, Version 1
 *
 * For now, this is just a simple poster application for CDM models
 * As a next step, this application can be extended to be able to read ASCII based
 * CDM models from files, and post them on the modelling bus.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 16.12.2025
 *
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
	cdm "github.com/erikproper/big-modelling-bus.go.v1/languages/cdm/cdm_v1_0_v1_0"
)

/*
 * Defining constants
 */

const (
	defaultIni = "config.ini" // Default configuration file name
)

/*
 * Defining flags
 */

var (
	configFlag      = flag.String("config", defaultIni, "Configuration file")               // Configuration file flag
	reportLevelFlag = flag.Int("reporting", generics.ProgressLevelBasic, "Reporting level") // Reporting level flag
)

/*
 * Pausing during posting. Just needed for testing purposes.
 */

func Pause() {
	fmt.Println("Press any key")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

/*
 * Main function
 */

func main() {
	// Parsing command line flags
	flag.Parse()

	// Creating the reporter
	reporter := generics.CreateReporter(*reportLevelFlag, generics.ReportError, generics.ReportProgress)

	// Loading the configuration
	configData := generics.LoadConfig(*configFlag, reporter)

	// Creating the Modelling Bus Connector
	ModellingBusConnector := connect.CreateModellingBusConnector(configData, reporter, connect.PostingOnly)

	//	ModellingBusConnector.DeleteEnvironment("experiment-12.10.2025")
	//	ModellingBusConnector.DeleteEnvironment("")

	//		ModellingBusConnector.PostRawArtefact("context", "golang", "test", "main.go")
	//		fmt.Println(ModellingBusConnector.GetRawArtefact("cdm-tester", "context", "golang", "test", "local.go"))
	//		fmt.Println(ModellingBusConnector.GetRawArtefact("cdm-tester", "context", "golang", "test", "local.go"))
	//		ModellingBusConnector.DeleteRawArtefact("context", "golang", "test.go")

	// Note that the 0001 is for local use. No issue to e.g. make this into 0001/02 to indicate version numbers
	CDMModellingBusPoster := cdm.CreateCDMPoster(ModellingBusConnector, "0001")

	CDMModel := cdm.CreateCDMModel(reporter)
	CDMModel.SetModelName("Empty university")

	fmt.Println("1) empty model")
	CDMModellingBusPoster.PostState(CDMModel)
	fmt.Println("Posted state")
	Pause()

	Student := CDMModel.AddConcreteIndividualType("Student")
	StudyProgramme := CDMModel.AddConcreteIndividualType("Study Programme")
	StudentName := CDMModel.AddQualityType("Student Name", "string")
	StudyProgrammeName := CDMModel.AddQualityType("Study Programme Name", "string")
	CDMModel.SetModelName("Basic university")

	fmt.Println("2) basic model")
	CDMModellingBusPoster.PostUpdate(CDMModel)
	fmt.Println("Posted update")
	Pause()

	fmt.Println("3) basic model")
	CDMModellingBusPoster.PostState(CDMModel)
	fmt.Println("Posted state")
	Pause()

	StudyProgrammeStudied := CDMModel.AddInvolvementType("studied by", StudyProgramme)
	StudentStudying := CDMModel.AddInvolvementType("studying", Student)
	Studies := CDMModel.AddRelationType("Studies", StudyProgrammeStudied, StudentStudying)
	CDMModel.AddRelationTypeReading(Studies, "", StudentStudying, "studies", StudyProgrammeStudied, "")
	CDMModel.AddRelationTypeReading(Studies, "", StudyProgrammeStudied, "studied by", StudentStudying, "")

	StudentReferred := CDMModel.AddInvolvementType("referred", Student)
	StudentNameReferring := CDMModel.AddInvolvementType("referring", StudentName)
	StudentNaming := CDMModel.AddRelationType("Student Naming", StudentReferred, StudentNameReferring)
	CDMModel.AddRelationTypeReading(StudentNaming, "", StudentReferred, "has", StudentNameReferring, "")
	CDMModel.AddRelationTypeReading(StudentNaming, "", StudentNameReferring, "of", StudentReferred, "")

	StudyProgrammeReferred := CDMModel.AddInvolvementType("referred", StudyProgramme)
	StudyProgrammeNameReferring := CDMModel.AddInvolvementType("referring", StudyProgrammeName)
	StudyProgrammeNaming := CDMModel.AddRelationType("Programme Naming", StudyProgrammeReferred, StudyProgrammeNameReferring)
	CDMModel.AddRelationTypeReading(StudyProgrammeNaming, "", StudyProgrammeReferred, "goes by", StudyProgrammeNameReferring, "")
	CDMModel.AddRelationTypeReading(StudyProgrammeNaming, "", StudyProgrammeNameReferring, "of", StudyProgrammeReferred, "")
	CDMModel.SetModelName("University")

	fmt.Println("4) larger model")
	CDMModellingBusPoster.PostUpdate(CDMModel)
	fmt.Println("Posted update")
	Pause()

	// Reference modes

	// CONSTRAINTS
	//
	// always do a push_model after a read from local FS!
	// push_model
	// push_update

	fmt.Println("5) final model")
	CDMModellingBusPoster.PostState(CDMModel)
}
