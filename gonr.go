package gonr

import (
	"github.com/neocortical/newrelic"
)

const defaultPluginName = "GoNR"
const defaultPluginGUID = "net.neocortical.newrelic.gonr"

type Config struct {
	Name    string
	GUID    string
	Runtime bool
	Memory  bool
	GC      bool
	HTTP    bool
}

var defaultConfig = Config{
	Runtime: true,
	Memory:  true,
	GC:      true,
	HTTP:    true,
}

// New creates a GoNR NewRelic client with the supplied license key
func New(license string) (*newrelic.Client, HttpMiddleware) {
	client := newrelic.New(license)
	plugin := NewPlugin(defaultConfig)

	client.AddPlugin(plugin)
	return client, nil
}

// NewWithConfig returns a GoNR NewRelic client configured with the supplied config
func NewWithConfig(license string, config Config) *newrelic.Client {
	client := newrelic.New(license)
	plugin := NewPlugin(config)

	client.AddPlugin(plugin)
	return client
}

// NewPlugin returns a GoNR NewRelic plugin (add to your own client) configured with the supplied config
func NewPlugin(config Config) *newrelic.Plugin {
	if config.Name == "" {
		config.Name = defaultPluginName
	}
	if config.GUID == "" {
		config.GUID = defaultPluginGUID
	}

	plugin := &newrelic.Plugin{
		Name: config.Name,
		GUID: config.GUID,
	}

	if config.Runtime {
		addRuntimeMetrics(plugin)
	}

	if config.GC {
		addGCMetrics(plugin)
	}

	if config.Memory {
		addMemoryMetrics(plugin)
	}

	// TODO: http

	return plugin
}
