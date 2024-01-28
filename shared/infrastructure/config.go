package infrastructure // todo config would be a better name OR move it to arrower.Config

// Config is a structure used for service configuration.
// It can be mapped from env variables or config files, e.g. by viper.
type Config struct {
	OrganisationName string `mapstructure:"organisation_name"`
	ApplicationName  string `mapstructure:"application_name"`
	InstanceName     string `mapstructure:"instance_name"`

	Debug bool `mapstructure:"debug"`

	Postgres Postgres `mapstructure:"postgres"`
	Web      Web      `mapstructure:"web"`
}

type (
	Postgres struct {
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		MaxConns int    `mapstructure:"max_conns"`
	}

	Web struct {
		Secret             []byte `mapstructure:"secret"`
		Port               int    `mapstructure:"port"`
		StatusEndpoint     bool   `mapstructure:"status_endpoint"`
		StatusEndpointPort int    `mapstructure:"status_endpoint_port"`
	}
)
