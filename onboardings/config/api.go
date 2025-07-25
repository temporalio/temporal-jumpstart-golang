package config

type APIConfig struct {
	URL  string `yaml:"url"`
	MTLS *struct {
		CertChainFile string `yaml:"cert_chain_file"`
		KayFile       string `yaml:"key_file"`
	} `yaml:"mtls"`
}
