package rest

import (
	"github.com/getevo/evo/v2"
)

// Controller represents a controller type.
type Controller struct{}

// Field represents a field in a data structure.
// It contains metadata about the field, such as its name, database name, type, default value, and whether it is a primary key.
type Field struct {
	Name      string `json:"label"`
	FieldName string `json:"-"`
	DBName    string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Default   string `json:"default,omitempty"`
	PK        bool   `json:"pk,omitempty"`
}

// Param represents a parameter used in the Resource struct.
type Param struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Primary bool   `json:"primary"`
}

// Models returns a collection of resources.
func (c Controller) Models(request *evo.Request) interface{} {
	return resources
}

// ORM is a method in the Controller struct that handles an ORM request.
// It expects a pointer to a Request object as a parameter and returns an interface{}
// It returns nil by default.
func (c Controller) ORM(request *evo.Request) interface{} {
	return nil
}
