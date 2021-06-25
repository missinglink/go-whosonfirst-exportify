package exportify

import (
	"flag"
	"context"
	"github.com/whosonfirst/go-whosonfirst-exportify/application"
)

func ExportifyApplication struct{
	application.Application
}

func NewExportifyApplication() (application.Application, error) {

	app := &ExportifyApplication{}
	return app, nil
}

func (app *ExportifyApplication) Run(ctx context.Context) error {
	fs := app.DefaultFlagSet(ctx)
	return app.RunWithFlagSet(ctx, fs)
}

func (app *ExportifyApplication) RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

}
