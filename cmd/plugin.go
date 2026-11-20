// cmd/plugin.go
package cmd

import (
	"fmt"
	"plugin"

	"sistemterdistribusi/internal/mr"
)

type (
	mapFunc    = func(string, string) []mr.KeyValue
	reduceFunc = func(string, []string) string
)

// loadPlugin opens a plugin file (for example wc.so).
// It returns the Map and Reduce functions from the plugin.
func loadPlugin(filename string) (mapFunc, reduceFunc, error) {
	p, err := plugin.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load plugin %v: %w", filename, err)
	}

	xmapf, err := p.Lookup("Map")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot find Map in %v: %w", filename, err)
	}
	mapf, ok := xmapf.(mapFunc)
	if !ok {
		return nil, nil, fmt.Errorf("Map in %v has the wrong type", filename)
	}

	xreducef, err := p.Lookup("Reduce")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot find Reduce in %v: %w", filename, err)
	}
	reducef, ok := xreducef.(reduceFunc)
	if !ok {
		return nil, nil, fmt.Errorf("Reduce in %v has the wrong type", filename)
	}

	return mapf, reducef, nil
}
