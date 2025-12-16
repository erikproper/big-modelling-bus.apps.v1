/*
 *
 * Module:      BIG Modelling Bus Apps, Version 1
 * Package:     Modelling Bus Apps
 * Application: LaTeX based PDF Renderer for CDM Models, Version 1
 *
 * This application listens to CDM model postings on the BIG Modelling Bus, and renders them as a PDF file using LaTeX.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 16.12.2025
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/erikproper/big-modelling-bus.go.v1/connect"
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
	cdm "github.com/erikproper/big-modelling-bus.go.v1/languages/cdm/cdm_v1_0_v1_0"
)

/*
 * Defining key constants
 */
const (
	defaultIni          = "config.ini"
	latexFileExtension  = ".tex"
	latexDefaultCommand = "pdflatex"
)

/*
 * Defining flags
 */
var (
	configFlag      = flag.String("config", defaultIni, "Configuration file")
	reportLevelFlag = flag.Int("reporting", generics.ProgressLevelBasic, "Reporting level")
	modelIDFlag     = flag.String("for_model", "", "Model ID to listen for")
	agentIDFlag     = flag.String("from_agent", "", "Agent ID to listen to")
)

/*
 * Defining the CDM model LaTeX writer
 */
type TCDMModelLaTeXWriter struct {
	cdm.TCDMModelListener // The CDM model listener

	latexFile    string // Name of the LaTeX file
	latexCommand string // Command to run LaTeX
	workFolder   string // Working folder

	LaTeXfile *os.File // The LaTeX file

	reporter *generics.TReporter // The Reporter to be used to report progress, errors, and panics
}

/*
 *  String constants for LaTeX formatting
 */
const (
	toAdd          = "{\\color{green} %s}"
	toDelete       = "{\\color{red} \\sout{\\sout{%s}}}"
	considerAdd    = "{\\color{lime} %s}"
	considerDelete = "{\\color{orange} \\sout{\\sout{%s}}}"
)

/*
 * Rendering elements with LaTeX formatting
 */

// Applying formatting
func ApplyFormatting(format, value string) string {
	if value == "" {
		return ""
	} else {
		return fmt.Sprintf(format, value)
	}
}

// Rendering model elements
func (l *TCDMModelLaTeXWriter) RenderElement(s func(cdm.TCDMModel) string) string {
	// Getting the current, updated, and considered model elements via the access function s
	current := s(l.CurrentModel)
	updated := s(l.UpdatedModel)
	considered := s(l.ConsideredModel)

	// Deciding on the formatting to apply
	if considered == updated {
		// No changes between the considered version and the updated version
		if updated == current {
			// No changes between the current version and the updated version
			return current
		} else {
			// Changes between the current version and the updated version
			return ApplyFormatting(toDelete, current) + ApplyFormatting(toAdd, updated)
		}
	} else {
		// Changes between the considered version and the updated version
		if updated == current {
			// No changes between the current version and the updated version
			return ApplyFormatting(considerDelete, updated) + ApplyFormatting(considerAdd, considered)
		} else {
			// Changes between the current version and the updated version
			return ApplyFormatting(toDelete, current) + ApplyFormatting(considerDelete, updated) + ApplyFormatting(considerAdd, considered)
		}
	}
}

// Render the model name
func (l *TCDMModelLaTeXWriter) RenderModelName() string {
	return l.RenderElement(func(m cdm.TCDMModel) string {
		return m.ModelName
	})
}

// Render the type name of the base type of an involvement type
func (l *TCDMModelLaTeXWriter) RenderTypeNameOfBaseTypeOfInvolvementType(involvementType string) string {
	return l.RenderElement(func(m cdm.TCDMModel) string {
		return m.TypeName[m.BaseTypeOfInvolvementType[involvementType]]
	})
}

// Render the domain name of a quality type
func (l *TCDMModelLaTeXWriter) RenderDomainNameOfQualityType(typeID string) string {
	return l.RenderElement(func(m cdm.TCDMModel) string {
		return m.DomainOfQualityType[typeID]
	})
}

// Render the type name
func (l *TCDMModelLaTeXWriter) RenderTypeName(typeID string) string {
	return l.RenderElement(func(m cdm.TCDMModel) string {
		return m.TypeName[typeID]
	})
}

// Render a relation type reading
func (l *TCDMModelLaTeXWriter) RenderRelationTypeReading(m cdm.TCDMModel, reading string) string {
	readingString := ""
	for involvementPosition, involvementType := range m.ReadingDefinition[reading].InvolvementTypes {
		if involvementPosition == 0 {
			readingString += m.ReadingDefinition[reading].ReadingElements[involvementPosition]
		}
		readingString += " " +
			m.TypeName[m.BaseTypeOfInvolvementType[involvementType]] +
			" $\\{$ " + m.TypeName[involvementType] + " $\\}$ " +
			m.ReadingDefinition[reading].ReadingElements[involvementPosition+1]
	}
	return strings.TrimSpace(readingString)
}

// Render the primary relation type reading
func (l *TCDMModelLaTeXWriter) RenderPrimaryRelationTypeReading(relationTypeID string) string {
	return l.RenderElement(func(m cdm.TCDMModel) string {
		return l.RenderRelationTypeReading(m, m.PrimaryReadingOfRelationType[relationTypeID])
	})
}

// Render a relation type reading
func (l *TCDMModelLaTeXWriter) RenderAlternativeRelationTypeReading(reading string) string {
	return l.RenderElement(func(m cdm.TCDMModel) string {
		return l.RenderRelationTypeReading(m, reading)
	})
}

/*
 * Writing LaTeX files
 */

// Writing formatted strings to the LaTeX file
func (l *TCDMModelLaTeXWriter) WriteLaTeX(format string, parameters ...any) {
	// Writing to the LaTeX file
	l.LaTeXfile.WriteString(fmt.Sprintf(format, parameters...))
}

// Writing types to the LaTeX file
func (l *TCDMModelLaTeXWriter) WriteTypesToLaTeX(sectionTitle string, types map[string]bool, writeTypeToLaTeX func(string)) {
	// Writing the types to the LaTeX file

	// Let's assume the list is empty, by default.
	empty := true
	for tpe, included := range types {
		if included {
			// Writing the type, if included
			if empty {
				// Writing the section header
				l.WriteLaTeX("\\section{%s}\n", sectionTitle)
				l.WriteLaTeX("\\begin{itemize}\n")
			} else {
				// Adding a new line between types
				l.WriteLaTeX("\n")
			}

			// Marking that the list is not empty
			empty = false

			// Writing the type itself
			writeTypeToLaTeX(tpe)
		}
	}

	// Closing the itemize environment, if needed
	if !empty {
		l.WriteLaTeX("\\end{itemize}\n")
		l.WriteLaTeX("\n")
	}
}

// Writing the model to a LaTeX file
func (l *TCDMModelLaTeXWriter) WriteModelToLaTeX() {
	// Creating the LaTeX file
	l.LaTeXfile, _ = os.Create(l.workFolder + "/" + l.latexFile + latexFileExtension)

	// Ensuring the LaTeX file is closed afterwards
	defer l.LaTeXfile.Close()

	// Writing the LaTeX file header
	l.WriteLaTeX("\\documentclass[a4paper]{article}\n")
	l.WriteLaTeX("\\usepackage{a4wide}\n")
	l.WriteLaTeX("\\usepackage{xcolor}\n")
	l.WriteLaTeX("\\usepackage{ulem}\n")
	l.WriteLaTeX("\n")
	l.WriteLaTeX("\\title{CDM Model: %s}\n", l.RenderModelName())
	l.WriteLaTeX("\\author{~~}\n")
	l.WriteLaTeX("\n")
	l.WriteLaTeX("\\begin{document}\n")
	l.WriteLaTeX("\\maketitle\n")
	l.WriteLaTeX("\n")

	// Writing the quality types to the LaTeX file
	l.WriteTypesToLaTeX("Quality types", l.QualityTypes(), func(qualityType string) {
		l.WriteLaTeX("    \\item {\\sf %s} with domain {\\sf %s}\n", l.RenderTypeName(qualityType), l.RenderDomainNameOfQualityType(qualityType))
	})

	// Writing the concrete individual types to the LaTeX file
	l.WriteTypesToLaTeX("Concrete individual types", l.ConcreteIndividualTypes(), func(concreteIndividualType string) {
		l.WriteLaTeX("    \\item {\\sf %s}\n", l.RenderTypeName(concreteIndividualType))
	})

	// Writing the relation types to the LaTeX file
	l.WriteTypesToLaTeX("Relation types", l.RelationTypes(), func(relationType string) {
		l.WriteLaTeX("    \\item {\\sf %s: $\\{$ ", l.RenderTypeName(relationType))

		// Writing the involvement types of the relation type
		sep := ""
		for involvementType, included := range l.InvolvementTypesOfRelationType(relationType) {
			if included {
				l.WriteLaTeX("%s%s %s", sep, l.RenderTypeNameOfBaseTypeOfInvolvementType(involvementType), l.RenderTypeName(involvementType))
				sep = "; "
			}
		}
		l.WriteLaTeX(" $\\}$}\n")

		// Writing the primary reading of the relation type
		if primaryRelationTypeReading := l.RenderPrimaryRelationTypeReading(relationType); primaryRelationTypeReading != "" {
			l.WriteLaTeX("\n")
			l.WriteLaTeX("          Primary reading:\n")
			l.WriteLaTeX("          \\begin{itemize}\n")
			l.WriteLaTeX("              \\item {\\sf %s}\n", primaryRelationTypeReading)
			l.WriteLaTeX("          \\end{itemize}\n")
		}

		// Writing the alternative readings of the relation type
		if len(l.AlternativeReadingsOfRelationType(relationType)) > 0 {
			l.WriteLaTeX("\n")
			l.WriteLaTeX("          Alternative reading(s):\n")
			l.WriteLaTeX("          \\begin{itemize}\n")
			readingPosition := 0
			for reading := range l.AlternativeReadingsOfRelationType(relationType) {
				if readingPosition > 0 {
					l.WriteLaTeX("\n")
				}
				readingPosition++
				l.WriteLaTeX("              \\item {\\sf %s}\n", l.RenderAlternativeRelationTypeReading(reading))
			}
			l.WriteLaTeX("          \\end{itemize}\n")
		}
	})

	// Writing the LaTeX file footer
	l.WriteLaTeX("\\end{document}\n")
}

// Creating the PDF file from the LaTeX file
func (l *TCDMModelLaTeXWriter) CreatePDF() {
	// Creating the PDF file using pdflatex

	// Set the LaTex command, which we ony need to run once for this application
	cmd := exec.Command("pdflatex", l.latexFile+latexFileExtension)

	// Setting the working directory
	cmd.Dir = l.workFolder

	// Running the command
	cmd.Run()
}

func CreateCDMLaTeXWriter(configData *generics.TConfigData, modelListener cdm.TCDMModelListener, reporter *generics.TReporter) TCDMModelLaTeXWriter {
	// Creating the CDM model LaTeX writer
	CDMModelLaTeXWriter := TCDMModelLaTeXWriter{}
	CDMModelLaTeXWriter.reporter = reporter
	CDMModelLaTeXWriter.TCDMModelListener = modelListener

	// Setting up the LaTeX writer based on the config data
	CDMModelLaTeXWriter.workFolder = configData.GetValue("", "work_folder").String()
	CDMModelLaTeXWriter.latexFile = configData.GetValue("", "latex").String()
	CDMModelLaTeXWriter.latexCommand = configData.GetValue("", "latex_command").StringWithDefault(latexDefaultCommand)

	// Returning the created LaTeX writer
	return CDMModelLaTeXWriter
}

// Updating the rendering based on the current model state
func (l *TCDMModelLaTeXWriter) UpdateRendering(message string) {
	// Reporting on the update
	l.reporter.Progress(generics.ProgressLevelBasic, "%s", message)

	// Writing the model to LaTeX and creating the PDF
	l.WriteModelToLaTeX()

	// Creating the PDF
	l.CreatePDF()
}

func (l *TCDMModelLaTeXWriter) ListenForModelPostings(agentID, modelID string) {
	// Listening for model state postings
	l.ListenForModelStatePostings(agentID, modelID, func() {
		l.UpdateRendering("Received state.")
	})

	// Listening for model update postings
	l.ListenForModelUpdatePostings(agentID, modelID, func() {
		l.UpdateRendering("Received update.")
	})

	// Listening for model considering postings
	l.ListenForModelConsideringPostings(agentID, modelID, func() {
		l.UpdateRendering("Received considered.")
	})
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

	// Validating agent ID flag
	if len(*agentIDFlag) == 0 {
		reporter.Error("No agent ID specified.")

		return
	}

	// Validating agent ID flag
	if len(*agentIDFlag) == 0 {
		reporter.Error("No agent ID specified.")

		return
	}

	// Validating model ID flag
	if len(*modelIDFlag) == 0 {
		reporter.Error("No model ID specified.")

		return
	}

	// Reporting progress
	reporter.Progress(generics.ProgressLevelBasic, "Starting LaTeX based PDF renderer for CDM models")
	reporter.Progress(generics.ProgressLevelBasic, "Listening for model ID '%s' from agent ID '%s'", *modelIDFlag, *agentIDFlag)

	// Note: the config data can be used to contain config data for different aspects
	configData := generics.LoadConfig(*configFlag, reporter)

	// Note: One ModellingBusConnector can be used for different models of different kinds.
	ModellingBusConnector := connect.CreateModellingBusConnector(configData, reporter, !connect.PostingOnly)

	// Creating the CDM model listener
	CDMModellingBusListener := cdm.CreateCDMListener(ModellingBusConnector, reporter)

	// Creating the CDM model LaTeX writer
	CDMLaTeXWriter := CreateCDMLaTeXWriter(configData, CDMModellingBusListener, reporter)

	// Setting up listening for model postings
	CDMLaTeXWriter.ListenForModelPostings(*agentIDFlag, *modelIDFlag)

	// Keeping the application running
	for {
		time.Sleep(1 * time.Second)
	}
}
