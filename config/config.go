// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type ETWConfig struct {
	Providers []Provider `config:"providers"`
}

type Provider struct {
	Id          string `config:"id"`
	SessionName string `config:"session_name"`
}
