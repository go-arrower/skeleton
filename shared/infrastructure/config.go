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
	OTEL     OTEL     `mapstructure:"otel"`
}

type (
	Postgres struct {
		User     string `json:"user"     mapstructure:"user"`
		Password Secret `json:"-"        mapstructure:"password"`
		Database string `json:"database" mapstructure:"database"`
		Host     string `json:"host"     mapstructure:"host"`
		Port     int    `json:"port"     mapstructure:"port"`
		SSLMode  string `json:"sslMode"  mapstructure:"ssl_mode"`
		MaxConns int    `json:"maxConns" mapstructure:"max_conns"`
	}

	Web struct {
		Hostname           string `json:"hostname" mapstructure:"hostname"`
		Port               int    `json:"port"     mapstructure:"port"`
		Secret             Secret `json:"-"        mapstructure:"secret"`
		StatusEndpoint     bool   `json:"-"        mapstructure:"status_endpoint"`
		StatusEndpointPort int    `json:"-"        mapstructure:"status_endpoint_port"`
	}

	OTEL struct {
		Host string `json:"host" mapstructure:"host"`
		Port int    `json:"port" mapstructure:"port"`
	}
)

// Secret is used to store sensitive data.
// It is masked should you output it somewhere.
//
// s := Secret("my-secret")
// fmt.Println(s)										=> output: ******
// t.Log(s)												=> output: ******
// logger.Info("", slog.Any("secret", s))				=> output: ******
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
