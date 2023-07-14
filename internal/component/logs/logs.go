package logs

import (
	"path/filepath"
	"strings"
)

type Type int

const (
	TypeGeneric Type = iota
	TypeNDJSON
)

func (t Type) String() string {
	switch t {
	case TypeGeneric:
		return "Generic"
	case TypeNDJSON:
		return "NDJSON"
	}

	return ""
}

type Component int

const (
	ComponentGeneric Component = iota
	ComponentAgent
	ComponentFilebeat
	ComponentMetricbeat
)

func (c Component) String() string {
	switch c {
	case ComponentGeneric:
		return "Generic"
	case ComponentAgent:
		return "Agent"
	case ComponentFilebeat:
		return "Filebeat"
	case ComponentMetricbeat:
		return "Metricbeat"
	}

	return ""
}

type Entry struct {
	Filename  string
	Type      Type
	Component Component
}

func GetType(filename string) Type {
	ext := filepath.Ext(filename)

	switch ext {
	case ".ndjson":
		return TypeNDJSON
	}

	return TypeGeneric
}

func GetComponent(filename string) Component {
	filename = filepath.Base(filename)

	if strings.HasPrefix(filename, "elastic-agent") {
		return ComponentAgent
	}
	if strings.HasPrefix(filename, "filebeat") {
		return ComponentFilebeat
	}
	if strings.HasPrefix(filename, "metricbeat") {
		return ComponentMetricbeat
	}

	return ComponentGeneric
}
