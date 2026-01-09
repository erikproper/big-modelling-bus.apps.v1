// Package plantuml provides a parser for structural PlantUML models
// with support for entities, attributes, methods, relationships,
// multiplicities, and basic constraint extraction.
package plantuml

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// -----------------------------
// Model definitions
// -----------------------------

// Model represents a parsed PlantUML model.
type Model struct {
	Entities      map[string]*Entity
	Relationships []*Relationship
	Constraints   []*Constraint
}

// Entity represents a class / entity / object.
type Entity struct {
	Name       string
	Attributes []Attribute
	Methods    []Method
}

// Attribute represents a class attribute.
type Attribute struct {
	Name string
	Type string
}

// Method represents a class method.
type Method struct {
	Name       string
	ReturnType string
}

// Relationship represents an association between two entities.
type Relationship struct {
	From string
	To   string
	Type string // e.g. "--", "<|--", "*--"

	// Multiplicities as written in PlantUML, e.g. "1", "0..*"
	FromMultiplicity string
	ToMultiplicity   string

	Label string
}

// Constraint represents a parsed constraint (e.g. unique, mandatory).
type Constraint struct {
	Kind   string // unique, mandatory, subset, etc.
	Target string // entity or role
	Expr   string // raw textual expression
}

// -----------------------------
// Parser
// -----------------------------

type Parser struct {
	scanner      *bufio.Scanner
	model        *Model
	currentClass *Entity
}

// NewParser creates a new PlantUML parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(r),
		model: &Model{
			Entities:      make(map[string]*Entity),
			Relationships: []*Relationship{},
			Constraints:   []*Constraint{},
		},
	}
}

// Parse reads the input and returns a parsed model.
func (p *Parser) Parse() (*Model, error) {
	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())

		// Ignore empty lines and directives
		if line == "" || strings.HasPrefix(line, "@") || strings.HasPrefix(line, "'") {
			continue
		}

		// End of class body
		if line == "}" {
			p.currentClass = nil
			continue
		}

		// Inside class body
		if p.currentClass != nil {
			if parseAttribute(line, p.currentClass) {
				continue
			}
			if parseMethod(line, p.currentClass) {
				continue
			}
		}

		// Entity declaration
		if parseEntity(line, p) {
			continue
		}

		// Relationship declaration (with multiplicities)
		if parseRelationship(line, p.model) {
			continue
		}

		// Constraint declaration
		if parseConstraint(line, p.model) {
			continue
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, err
	}

	return p.model, nil
}

// -----------------------------
// Parsing helpers
// -----------------------------

var entityRegex = regexp.MustCompile(`^(class|entity|object)\s+(\w+)\s*\{?$`)

func parseEntity(line string, p *Parser) bool {
	matches := entityRegex.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	name := matches[2]
	entity := &Entity{Name: name}
	p.model.Entities[name] = entity
	p.currentClass = entity
	return true
}

var attributeRegex = regexp.MustCompile(`^(\w+)\s*:\s*(\w+)$`)

func parseAttribute(line string, e *Entity) bool {
	matches := attributeRegex.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	e.Attributes = append(e.Attributes, Attribute{
		Name: matches[1],
		Type: matches[2],
	})
	return true
}

var methodRegex = regexp.MustCompile(`^(\w+)\(.*\)\s*:\s*(\w+)$`)

func parseMethod(line string, e *Entity) bool {
	matches := methodRegex.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	e.Methods = append(e.Methods, Method{
		Name:       matches[1],
		ReturnType: matches[2],
	})
	return true
}

// Supports: A "1" -- "0..*" B : label
var relationRegex = regexp.MustCompile(
	`^(\w+)\s*("[^"]+")?\s+([-.o*<|]+)\s*("[^"]+")?\s+(\w+)(\s*:\s*(.+))?$`,
)

func parseRelationship(line string, model *Model) bool {
	matches := relationRegex.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	rel := &Relationship{
		From:             matches[1],
		FromMultiplicity: strings.Trim(matches[2], "\""),
		Type:             matches[3],
		ToMultiplicity:   strings.Trim(matches[4], "\""),
		To:               matches[5],
		Label:            matches[7],
	}

	model.Relationships = append(model.Relationships, rel)
	return true
}

var constraintRegex = regexp.MustCompile(`^constraint\s+(\w+)\s+on\s+(\w+)\s*:\s*(.+)$`)

func parseConstraint(line string, model *Model) bool {
	matches := constraintRegex.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	model.Constraints = append(model.Constraints, &Constraint{
		Kind:   matches[1],
		Target: matches[2],
		Expr:   matches[3],
	})
	return true
}

// -----------------------------
// Utility
// -----------------------------

func (m *Model) DebugPrint() {
	fmt.Println("Entities:")
	for _, e := range m.Entities {
		fmt.Println(" -", e.Name)
		for _, a := range e.Attributes {
			fmt.Printf("    attr %s : %s\n", a.Name, a.Type)
		}
		for _, m := range e.Methods {
			fmt.Printf("    method %s() : %s\n", m.Name, m.ReturnType)
		}
	}

	fmt.Println("Relationships:")
	for _, r := range m.Relationships {
		fmt.Printf(
			" - %s \"%s\" %s \"%s\" %s : %s\n",
			r.From,
			r.FromMultiplicity,
			r.Type,
			r.ToMultiplicity,
			r.To,
			r.Label,
		)
	}

	fmt.Println("Constraints:")
	for _, c := range m.Constraints {
		fmt.Printf(" - %s on %s : %s\n", c.Kind, c.Target, c.Expr)
	}
}
