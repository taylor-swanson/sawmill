package config

type Type int

const (
	TypeGeneric Type = iota
	TypeAgent
	TypeAgentPolicy
	TypeEndpoint
	TypeFilebeat
	TypeFleetMonitoring
	TypeMetricbeat
)

func (t Type) String() string {
	switch t {
	case TypeGeneric:
		return "Generic"
	case TypeAgent:
		return "Agent"
	case TypeAgentPolicy:
		return "Agent Policy"
	case TypeEndpoint:
		return "Endpoint"
	case TypeFilebeat:
		return "Filebeat"
	case TypeFleetMonitoring:
		return "Fleet Monitoring"
	case TypeMetricbeat:
		return "Metricbeat"
	}

	return ""
}

type Entry struct {
	Filename string
	Type     Type
}
