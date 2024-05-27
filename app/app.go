package app

import "github.com/getevo/evo/v2/lib/log"

type Application interface {
	Register() error
	Router() error
	WhenReady() error
	Name() string
}

type App struct {
	apps []Application
}

func New() *App {
	return &App{}
}
func (a *App) Register(applications ...Application) *App {
	a.apps = append(a.apps, applications...)
	return a
}

func (a *App) Run() *App {
	for _, app := range a.apps {
		if err := app.Register(); err != nil {
			log.Fatalf("Can't start application Register() %s: %s", app.Name(), err)
		}
		if err := app.Router(); err != nil {
			log.Fatalf("Can't start application Router() %s: %s", app.Name(), err)
		}
	}
	for _, app := range a.apps {
		if err := app.WhenReady(); err != nil {
			log.Fatalf("Can't start application WhenReady() %s: %s", app.Name(), err)
		}
	}

	return a
}
