package domain

type Suite struct {
	DefaultContentType string            `json:"default_content_type" yaml:"default_content_type"`
	GlobalHeaders      map[string]string `json:"global_headers" yaml:"global_headers"`
	Endpoints          []Endpoint        `json:"endpoints" yaml:"endpoints"`
}
