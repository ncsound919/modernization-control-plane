package hl7

import (
	"fmt"
	"strings"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/sidecar-gateway/internal/models"
)

// FHIRPatient is a simplified FHIR R4 Patient resource.
type FHIRPatient struct {
	ResourceType string        `json:"resourceType"`
	ID           string        `json:"id"`
	Meta         FHIRMeta      `json:"meta"`
	Identifier   []FHIRIdent   `json:"identifier"`
	Name         []FHIRHumanName `json:"name"`
	Gender       string        `json:"gender,omitempty"`
	BirthDate    string        `json:"birthDate,omitempty"`
}

type FHIRMeta struct {
	Profile     []string `json:"profile"`
	LastUpdated string   `json:"lastUpdated"`
}

type FHIRIdent struct {
	System string `json:"system"`
	Value  string `json:"value"`
}

type FHIRHumanName struct {
	Use    string   `json:"use"`
	Family string   `json:"family"`
	Given  []string `json:"given"`
}

// TransformResult wraps the FHIR output and transformation metadata.
type TransformResult struct {
	Source    string       `json:"source_version"`
	Target    string       `json:"target_version"`
	Patient   *FHIRPatient `json:"resource"`
	Warnings  []string     `json:"warnings,omitempty"`
}

// Adapter transforms HL7 v2 ADT messages into FHIR R4 resources.
type Adapter struct{}

// New returns a new HL7 Adapter.
func New() *Adapter { return &Adapter{} }

// Transform parses a basic HL7 v2 ADT^A01 message and maps it to a FHIR R4
// Patient resource. Only MSH, PID, and PV1 segments are processed; unknown
// segments are skipped with a warning.
func (a *Adapter) Transform(msg models.HL7Message) (*TransformResult, error) {
	if msg.Content == "" {
		return nil, fmt.Errorf("hl7 message content is required")
	}

	patient := &FHIRPatient{
		ResourceType: "Patient",
		ID:           fmt.Sprintf("hl7-%d", time.Now().UnixMilli()),
		Meta: FHIRMeta{
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Patient"},
			LastUpdated: time.Now().UTC().Format(time.RFC3339),
		},
	}

	result := &TransformResult{
		Source:  "HL7v2",
		Target:  "FHIR-R4",
		Patient: patient,
	}

	// Normalize line endings (HL7 uses CR as segment terminator).
	content := strings.ReplaceAll(msg.Content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	segments := strings.Split(content, "\n")

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if len(seg) < 3 {
			continue
		}

		switch seg[:3] {
		case "MSH":
			// MSH already validated — message control ID could be extracted
			// from field 10, but we skip for brevity.
		case "PID":
			parsePID(seg, patient, result)
		case "PV1":
			// Visit information — not mapped to Patient in this stub.
		default:
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("segment %s not mapped", seg[:3]))
		}
	}

	// Provide sensible defaults if the message was minimal / stub.
	if len(patient.Name) == 0 {
		patient.Name = []FHIRHumanName{{
			Use:    "official",
			Family: "Doe",
			Given:  []string{"John"},
		}}
		result.Warnings = append(result.Warnings, "PID-5 missing; default name applied")
	}
	if len(patient.Identifier) == 0 {
		patient.Identifier = []FHIRIdent{{
			System: "urn:oid:2.16.840.1.113883.4.1",
			Value:  "UNKNOWN",
		}}
	}

	return result, nil
}

// parsePID extracts PID-3 (patient ID), PID-5 (name), PID-7 (DOB) and
// PID-8 (gender) from a PID segment string.
func parsePID(seg string, patient *FHIRPatient, result *TransformResult) {
	fields := strings.Split(seg, "|")

	// PID-3: Patient Identifier List
	if len(fields) > 3 && fields[3] != "" {
		patient.Identifier = append(patient.Identifier, FHIRIdent{
			System: "urn:oid:2.16.840.1.113883.4.1",
			Value:  strings.Split(fields[3], "^")[0],
		})
	}

	// PID-5: Patient Name (Family^Given)
	if len(fields) > 5 && fields[5] != "" {
		parts := strings.Split(fields[5], "^")
		name := FHIRHumanName{Use: "official"}
		if len(parts) > 0 {
			name.Family = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			name.Given = []string{parts[1]}
		}
		patient.Name = append(patient.Name, name)
	}

	// PID-7: Date of Birth (YYYYMMDD)
	if len(fields) > 7 && len(fields[7]) == 8 {
		d := fields[7]
		patient.BirthDate = fmt.Sprintf("%s-%s-%s", d[0:4], d[4:6], d[6:8])
	}

	// PID-8: Administrative Sex
	if len(fields) > 8 {
		switch strings.ToUpper(fields[8]) {
		case "M":
			patient.Gender = "male"
		case "F":
			patient.Gender = "female"
		case "O":
			patient.Gender = "other"
		default:
			patient.Gender = "unknown"
		}
	}
}
