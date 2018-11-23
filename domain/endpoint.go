package domain

//go:generate slices Endpoint all

// Endpoint represents an HTTP endpoint to be mimicked by FONY
type Endpoint struct {
	URL           string     `json:"url" yaml:"url"`
	Method        string     `json:"method" yaml:"method"`
	Responses     []Response `json:"responses" yaml:"responses"`
	RunSequential bool       `json:"run_sequential" yaml:"run_sequential"`
}
