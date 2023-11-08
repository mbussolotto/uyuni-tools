package types

type ImageFlags struct {
	Name       string `mapstructure:"image"`
	Tag        string `mapstructure:"tag"`
	PullPolicy string `mapstructure:"pullPolicy"`
}
