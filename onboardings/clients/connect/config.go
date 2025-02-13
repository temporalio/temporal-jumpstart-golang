package connect

type Config struct {
	Port int `envconfig:"default=9000"`
}

func (c *Config) Prefix() string {
	return "grpc"
}
