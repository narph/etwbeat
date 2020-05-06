// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type ETWConfig struct {
	Sessions []Session `config:"sessions"`
}

type Session struct {
	Providers  []string `config:"provider_ids"`
	Name       string   `config:"name"`
	TraceLevel string   `config:"trace_level"`
}
