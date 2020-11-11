package suna

type Config struct {
	SecretKey []byte `toml:"secret-cHKey"`
}

func DefaultConfig() *Config {
	v := &Config{}
	return v
}
