package domain

//go:generate slices Response all

// Response represents an HTTP response from FONY
type Response struct {
	Headers    map[string]string `json:"headers" yaml:"headers"`
	Payload    interface{}       `json:"payload" yaml:"payload"`
	StatusCode int               `json:"status_code" yaml:"status_code"`
	Delay      float64           `json:"delay" yaml:"delay"`
	IsJSON     bool              `json:"-" yaml:"-"`
}
