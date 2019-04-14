package tlparser

// Schema is object containing types, classes and functions decribed in schema.
type Schema struct {
	Types     []*Type     `json:"types"`
	Classes   []*Class    `json:"classes"`
	Functions []*Function `json:"functions"`
}

// Type contains info about schema type.
type Type struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Class       string      `json:"class"`
	Properties  []*Property `json:"properties"`
}

// Class contains info about schema class.
type Class struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FunctionType contains info about schema function type.
type FunctionType int

// FunctionType constants.
const (
	FunctionTypeUnknown FunctionType = iota
	FunctionTypeCommon
	FunctionTypeUser
	FunctionTypeBot
)

// Function contains info about schema function.
type Function struct {
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Class         string       `json:"class"`
	Properties    []*Property  `json:"properties"`
	IsSynchronous bool         `json:"is_synchronous"`
	Type          FunctionType `json:"type"`
}

// Property contains info about schema property.
type Property struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
