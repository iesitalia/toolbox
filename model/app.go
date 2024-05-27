package model

import (
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db"
)

// Register is a function that registers models and callbacks in the database.
// It uses the `db.UseModel` function to set up the models for the database tables.
// The `Callback` struct is used to define a callback function that will be called after certain database operations.
// The `Register` function creates three callback functions (`OnCreate`, `OnUpdate`, and `OnDelete`) and registers them for the corresponding database operations (create, update, and
func Register() {
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

}
