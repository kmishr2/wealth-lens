package finance

type MetricDefinition struct {
	Name           string   `json:"name"`
	Formula        string   `json:"formula"`
	Assumptions    []string `json:"assumptions"`
	RequiredInputs []string `json:"required_inputs"`
	Explanation    string   `json:"explanation"`
}
