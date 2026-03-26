package models

// Adapter represents a registered legacy protocol adapter.
type Adapter struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Protocol string `json:"protocol"`
}

// Contract represents a versioned stable API contract exposed to consumers.
type Contract struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Spec    string `json:"spec"`
	Backend string `json:"backend"`
}

// ProxyRequest is the payload for proxying a request to a backend service.
type ProxyRequest struct {
	Service string            `json:"service"`
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// GatewayMetrics holds aggregate observability data for the gateway.
type GatewayMetrics struct {
	RequestCount  int64   `json:"request_count"`
	ErrorRate     float64 `json:"error_rate"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
}

// COBOLJob describes a COBOL batch job execution request.
type COBOLJob struct {
	ProgramName string            `json:"program_name"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	JCLTemplate string            `json:"jcl_template,omitempty"`
}

// HL7Message represents an inbound HL7 v2 message for transformation.
type HL7Message struct {
	Version     string `json:"version"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
}
