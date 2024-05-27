package model

import (
	"gorm.io/gorm"
	"reflect"
)

// Callback represents a callback function that can be registered to be executed
// after certain events occur in the database.
type Callback struct {
}

// OnModify is a callback method that is triggered after a modify operation (insert, update, delete) on the database.
// It checks if the database operation was successful and if the schema is not nil.
// If the schema represents a struct, it determines the action based on the build clauses.
// It then calls the corresponding "OnCreate", "OnUpdate", or "OnDelete" method on each field of the struct that has the action as a method.
// The method is called with two parameters: the db object and the reflect value of the schema.
func (c Callback) OnModify(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil {
		if db.Statement.ReflectValue.Kind() == reflect.Struct {
			var action = ""
			switch db.Statement.BuildClauses[0] {
			case "INSERT":
				action = "OnCreate"
			case "UPDATE":
				action = "OnUpdate"
			case "DELETE":
				action = "OnDelete"
			}
			if action == "" {
				return
			}
			for i := 0; i < db.Statement.ReflectValue.NumField(); i++ {
				var ref = db.Statement.ReflectValue.Field(i)
				if ref.CanAddr() {
					var m = ref.Addr().MethodByName(action)
					if m.IsValid() {
						m.Call([]reflect.Value{reflect.ValueOf(db), reflect.ValueOf(db.Statement.ReflectValue)})
					}
				}

			}

		}

	}
}
