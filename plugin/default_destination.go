package plugin

import (
	"github.com/bluemedora/bplogagent/entry"
)

type DefaultDestinationConfig struct {
	PluginID   string `json:"id" yaml:"id" mapstructure:"id"`
	BufferSize uint
}

func (c DefaultDestinationConfig) Build() DefaultDestination {
	bufferSize := c.BufferSize
	if bufferSize == 0 {
		bufferSize = 100
	}
	return DefaultDestination{
		config: c,
		input:  make(chan entry.Entry, bufferSize),
	}
}

func (c DefaultDestinationConfig) ID() string {
	return c.PluginID
}

type DefaultDestination struct {
	config DefaultDestinationConfig
	input  chan entry.Entry
}

func (s *DefaultDestination) Input() chan entry.Entry {
	return s.input
}

func (s *DefaultDestination) ID() string {
	return s.config.ID()
}