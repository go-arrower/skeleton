package infrastructure // todo config would be a better name OR move it to arrower.Config

import "encoding/json"

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
		User     string `mapstructure:"user" json:"user"`
		Password Secret `mapstructure:"password" json:"-"`
		Database string `mapstructure:"database" json:"database"`
		Host     string `mapstructure:"host" json:"host"`
		Port     int    `mapstructure:"port" json:"port"`
		MaxConns int    `mapstructure:"max_conns" json:"maxConns"`
	}

	Web struct {
		Port               int    `mapstructure:"port" json:"port"`
		Hostname           string `mapstructure:"hostname" json:"hostname"`
		Secret             []byte `mapstructure:"secret" json:"-"` // todo use Secret type
		StatusEndpoint     bool   `mapstructure:"status_endpoint" json:"-"`
		StatusEndpointPort int    `mapstructure:"status_endpoint_port" json:"-"`
	}
)

// Secret is used to store sensitive data.
// It is masked should you output it somewhere.
//
// s := Secret("my-secret")
// fmt.Println(s)							=> output: ******
// t.Log(s)									=> output: ******
// logger.Info("", slog.Any("secret", s))	=> output: ******
// logger.Info("", slog.String("secret", string(s)))	=> DON'T DO THIS! The secret will be exposed.
type Secret string

func (p Secret) String() string {
	return "******"
}

func (p Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p Secret) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}
