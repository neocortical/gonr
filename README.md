# GoNR: Go NewRelic plugin client

GoNR is a customizable NewRelic plugin written in Go. It allows you to get insight into your Go app from the NewRelic dashboard. 

This project (and its underlying libraries) was aided and inpired by [yvasiyarov's](https://github.com/yvasiyarov) [GoRelic plugin](https://github.com/yvasiyarov/gorelic). I chose to rewrite GoRelic with a few ideas for improvement in mind: 

* Easy, intuitive customizability
* Code clarity/package composability
* Unit testing
* Finer-grained control over metrics

# Installation

```go
go get github.com/neocortical/gonr
```

# Use

### Default (Use the plugin dashboard I've set up)
```go
import "github.com/neocortical/gonr"

func main() {
  gonrAgent := gonr.New(gonr.DefaultConfig().WithLicense("abc123"))
  gonrAgent.Run()
}

```

### Custom (Create your own plugin dashboard and/or add your own custom metrics)
```go
import "github.com/neocortical/gonr"

func main() {
  gonrAgent := gonr.New(gonr.Config{
		Name:    "My App",
		GUID:    "com.example.myapp.mynrplugin",
		License: "abc123",
	})
	
	gonrAgent.GetPlugin().AddMetric(myAwesomeMetric)
	
	gonrAgent.Run()
}
```

See my [newrelic plugin library](https://github.com/neocortical/newrelic) for info on how to easily create custom metrics.

