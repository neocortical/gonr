package gonr

import (
	"github.com/neocortical/newrelic"
)

const defaultPluginName = "GoNR"
const defaultPluginGUID = "net.neocortical.newrelic.gonr"

type Config struct {
	Name           string
	GUID           string
	License        string
	ExcludeRuntime bool
	ExcludeMemory  bool
	ExcludeGC      bool
}

var defaultConfig = Config{
	Name: defaultPluginName,
	GUID: defaultPluginGUID,
}

// WithLicense applies an API license key to a config. Ex: DefaultConfig().WithLicense("abc123")
func (c Config) WithLicense(license string) Config {
	c.License = license
	return c
}

// DefaultConfig returns the default GoNR config (all components, default Name and GUID)
func DefaultConfig() Config {
	return defaultConfig
}

type GonrAgent interface {
	Run()
	Client() *newrelic.Client
	Plugin() *newrelic.Plugin
}

// New creates a GoNR NewRelic client with the supplied license key
func New(config Config) GonrAgent {
	client := newrelic.New(config.License)
	plugin := newPlugin(config)
	client.AddPlugin(plugin)

	result := &gonrAgent{
		client: client,
		plugin: plugin,
	}
	return result
}

type gonrAgent struct {
	client *newrelic.Client
	plugin *newrelic.Plugin
}

// Run the underlying client
func (ga *gonrAgent) Run() {
	ga.client.Run()
}

// Client retrieves the underlying client to allow for customization
func (ga *gonrAgent) Client() *newrelic.Client {
	return ga.client
}

// Plugin retrives the underlying GoNR plugin object, allowing for customization
func (ga *gonrAgent) Plugin() *newrelic.Plugin {
	return ga.plugin
}

// newPlugin returns a NewRelic plugin configured for GoNR
func newPlugin(config Config) *newrelic.Plugin {
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

	if !config.ExcludeRuntime {
		addRuntimeMetrics(plugin)
	}

	if !config.ExcludeGC {
		addGCMetrics(plugin)
	}

	if !config.ExcludeMemory {
		addMemoryMetrics(plugin)
	}

	return plugin
}
