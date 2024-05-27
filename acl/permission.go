package acl

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/db/schema"
	"strings"
)

// Apps is a map variable that stores instances of the App struct.
var Apps = map[string]*App{}

// permissions represents a mapping of permission keys to Permission objects.
var permissions = map[string]*Permission{}

type Permission struct {
	ID          string `gorm:"column:id;size:64;primaryKey" json:"id"`
	App         string `gorm:"column:app;size:64;fk:permission_app" json:"app"`
	Key         string `gorm:"column:key;size:64" json:"key"`
	Name        string `gorm:"column:name;size:64" json:"name"`
	Description string `gorm:"column:description;size:64" json:"description"`
}

// TableName returns the name of the table for the Permission struct.
func (Permission) TableName() string {
	return "permission"
}

// App is a struct representing an application with its properties and permissions.
type App struct {
	App         string       `gorm:"column:app;size:64;primaryKey" json:"app"`
	Name        string       `gorm:"column:name;size:64" json:"name"`
	Description string       `gorm:"column:description;size:64" json:"description"`
	Permissions []Permission `gorm:"-" json:"permissions"`
}

// TableName returns the name of the table for the App model.
func (App) TableName() string {
	return "permission_app"
}

// Migration is a method that returns a list of schema migrations for the Permission type based on the provided version.
func (t Permission) Migration(version string) []schema.Migration {
	var migrations []schema.Migration
	for _, app := range Apps {
		var t App
		if db.Where("app = ?", app.Name).Take(&t).RowsAffected == 0 {
			migrations = append(migrations, schema.Migration{
				Query:   fmt.Sprintf("INSERT IGNORE INTO %s (`id`,`app`,`name`,`description`) VALUES ('%s','%s','%s')", t.TableName(), app.App, app.Name, app.Description),
				Version: "*",
			})
		}
	}
	for _, permission := range permissions {
		var t Permission
		if db.Where("`app` = ? AND `key` = ?", permission.App, permission.Key).Take(&t).RowsAffected == 0 {
			migrations = append(migrations, schema.Migration{
				Query:   fmt.Sprintf("INSERT IGNORE INTO %s (`id`,`app`,`key`,`name`,`description`) VALUES ('%s','%s','%s','%s','%s')", t.TableName(), permission.App+"."+permission.Key, permission.App, permission.Key, permission.Name, permission.Description),
				Version: "*",
			})
		}
	}

	return migrations
}

// SetPermission sets the permissions for an app by updating the `apps` and `permissions` maps.
func SetPermission(app *App) {
	Apps[app.App] = app
	for idx, perm := range app.Permissions {
		app.Permissions[idx].App = app.App
		permissions[strings.ToUpper(perm.App+"."+perm.Key)] = &app.Permissions[idx]
	}
}
