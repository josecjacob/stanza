package transformer

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/bluemedora/bplogagent/entry"
	"github.com/bluemedora/bplogagent/plugin"
	"github.com/bluemedora/bplogagent/plugin/helper"
	"go.uber.org/zap"
)

func init() {
	plugin.Register("router", &RouterPluginConfig{})
}

type RouterPluginConfig struct {
	helper.BasicConfig `yaml:",inline"`
	Routes             []*RouterPluginRouteConfig `json:"routes" yaml:"routes"`
}

type RouterPluginRouteConfig struct {
	Expression string           `json:"expr"   yaml:"expr"`
	OutputIDs  helper.OutputIDs `json:"output" yaml:"output"`
}

func (c RouterPluginConfig) Build(context plugin.BuildContext) (plugin.Plugin, error) {
	basicPlugin, err := c.BasicConfig.Build(context)
	if err != nil {
		return nil, err
	}

	routes := make([]*RouterPluginRoute, 0, len(c.Routes))
	for _, routeConfig := range c.Routes {
		compiled, err := expr.Compile(routeConfig.Expression, expr.AsBool(), expr.AllowUndefinedVariables())
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression '%s': %w", routeConfig.Expression, err)
		}
		route := RouterPluginRoute{
			Expression: compiled,
			OutputIDs:  routeConfig.OutputIDs,
		}
		routes = append(routes, &route)
	}

	routerPlugin := &RouterPlugin{
		BasicPlugin: basicPlugin,
		routes:      routes,
	}

	return routerPlugin, nil
}

func (c *RouterPluginConfig) SetNamespace(namespace string, exclusions ...string) {
	c.BasicConfig.SetNamespace(namespace, exclusions...)
	for _, route := range c.Routes {
		for i, outputID := range route.OutputIDs {
			if helper.CanNamespace(outputID, exclusions) {
				route.OutputIDs[i] = helper.AddNamespace(outputID, namespace)
			}
		}
	}
}

type RouterPlugin struct {
	helper.BasicPlugin
	routes []*RouterPluginRoute
}

type RouterPluginRoute struct {
	Expression    *vm.Program
	OutputIDs     helper.OutputIDs
	OutputPlugins []plugin.Plugin
}

func (p *RouterPlugin) CanProcess() bool {
	return true
}

func (p *RouterPlugin) Process(ctx context.Context, entry *entry.Entry) error {
	env := map[string]interface{}{
		"$": entry.Record,
	}

	for _, route := range p.routes {
		matches, err := vm.Run(route.Expression, env)
		if err != nil {
			p.Warnw("Running expression returned an error", zap.Error(err))
			continue
		}

		// we compile the expression with "AsBool", so this should be safe
		if matches.(bool) {
			for _, output := range route.OutputPlugins {
				_ = output.Process(ctx, entry)
			}
			break
		}
	}

	return nil
}

func (p *RouterPlugin) CanOutput() bool {
	return true
}

// Outputs will return all connected plugins.
func (p *RouterPlugin) Outputs() []plugin.Plugin {
	outputs := make([]plugin.Plugin, 0, len(p.routes))
	for _, route := range p.routes {
		outputs = append(outputs, route.OutputPlugins...)
	}
	return outputs
}

// SetOutputs will set the outputs of the router plugin.
func (p *RouterPlugin) SetOutputs(plugins []plugin.Plugin) error {
	for _, route := range p.routes {
		outputPlugins, err := p.findPlugins(plugins, route.OutputIDs)
		if err != nil {
			return fmt.Errorf("failed to set outputs on route: %s", err)
		}
		route.OutputPlugins = outputPlugins
	}
	return nil
}

// findPlugins will find a subset of plugins from a collection.
func (p *RouterPlugin) findPlugins(plugins []plugin.Plugin, pluginIDs []string) ([]plugin.Plugin, error) {
	result := make([]plugin.Plugin, 0)
	for _, pluginID := range pluginIDs {
		plugin, err := p.findPlugin(plugins, pluginID)
		if err != nil {
			return nil, err
		}
		result = append(result, plugin)
	}
	return result, nil
}

// findPlugin will find a plugin from a collection.
func (p *RouterPlugin) findPlugin(plugins []plugin.Plugin, pluginID string) (plugin.Plugin, error) {
	for _, plugin := range plugins {
		if plugin.ID() == pluginID {
			return plugin, nil
		}
	}
	return nil, fmt.Errorf("plugin %s does not exist", pluginID)
}