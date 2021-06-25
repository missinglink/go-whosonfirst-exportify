package application

type Application interface {
	Run(context.Context) error
	RunWithFlagSet(context.Context, *flag.FlagSet) error
}
