package rest

import (
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db/schema"
)

// PREFIX specifies the prefix for API routes in the admin panel.
var PREFIX = "/admin"

type App struct {
}

// Register registers all the resources and sets up the router for the application.
// For each model in `schema.Models`, it attaches a resource using the `AttachResource` method.
func (a App) Register() error {
	resources = map[string]*Resource{}

	for idx := range schema.Models {
		var model = schema.Models[idx]
		AttachResource(&model)
	}
	return nil
}

func (a App) Router() error {
	var controller = Controller{}
	evo.Get(PREFIX+"/rest/orm", controller.ORM)
	evo.Get(PREFIX+"/rest/models", controller.Models)
	return nil
}

func (a App) WhenReady() error {
	return nil
}

func (a App) Name() string {
	return "rest"
}
