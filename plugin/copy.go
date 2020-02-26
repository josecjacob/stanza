package plugin

import (
	"fmt"
	"sync"

	"github.com/bluemedora/bplogagent/entry"
	"go.uber.org/zap"
)

func init() {
	RegisterConfig("copy", &CopyConfig{})
}

type CopyConfig struct {
	DefaultPluginConfig
	DefaultInputterConfig
	Outputs []PluginID
	Field   string
}

func (c CopyConfig) Build(logger *zap.SugaredLogger) (Plugin, error) {
	defaultPlugin, err := c.DefaultPluginConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build default plugin: %s", err)
	}

	defaultInputter, err := c.DefaultInputterConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build default plugin: %s", err)
	}

	plugin := &CopyPlugin{
		DefaultPlugin:   defaultPlugin,
		DefaultInputter: defaultInputter,
		config:          c,
		input:           make(EntryChannel, c.BufferSize), // TODO default buffer size
		SugaredLogger:   logger.With("plugin_type", "copy", "plugin_id", c.PluginID),
	}

	return plugin, nil
}

type CopyPlugin struct {
	DefaultPlugin
	DefaultInputter

	outputs map[PluginID]EntryChannel
	input   EntryChannel
	config  CopyConfig
	*zap.SugaredLogger
}

func (s *CopyPlugin) Start(wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()
		for {
			entry, ok := <-s.input
			if !ok {
				return
			}

			for _, output := range s.outputs {
				// TODO should we block if one output can't keep up?
				output <- copyEntry(entry)
			}
		}
	}()

	return nil
}

func (s *CopyPlugin) SetOutputs(inputRegistry map[PluginID]EntryChannel) error {
	outputs := make(map[PluginID]EntryChannel, len(s.config.Outputs))
	for _, outputID := range s.config.Outputs {
		output, ok := inputRegistry[outputID]
		if !ok {
			return fmt.Errorf("no plugin with ID %v found", outputID)
		}

		outputs[outputID] = output
	}

	s.outputs = outputs
	return nil
}

func (s *CopyPlugin) Outputs() map[PluginID]EntryChannel {
	return s.outputs
}

func copyEntry(e entry.Entry) entry.Entry {
	newEntry := entry.Entry{}
	newEntry.Timestamp = e.Timestamp
	newEntry.Record = copyMap(e.Record)

	return newEntry
}