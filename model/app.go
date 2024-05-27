package model

import (
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db"
)

type App struct {
}

// Register registers all the resources and sets up the router for the application.
// For each model in `schema.Models`, it attaches a resource using the `AttachResource` method.
func (a App) Register() error {
	db.UseModel(TagEntity{}, TagList{})

	var callback Callback
	var dbo = evo.GetDBO()
	err := dbo.Callback().Create().After("*").Register("tag:create", callback.OnModify)
	if err != nil {
		panic(err)
	}
	err = dbo.Callback().Update().After("*").Register("tag:update", callback.OnModify)
	if err != nil {
		panic(err)
	}
	err = dbo.Callback().Delete().After("*").Register("tag:delete", callback.OnModify)
	if err != nil {
		panic(err)
	}
	return nil
}

func (a App) Router() error {
	return nil
}

func (a App) WhenReady() error {
	return nil
}

func (a App) Name() string {
	return "rest"
}
