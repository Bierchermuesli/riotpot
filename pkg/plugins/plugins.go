package plugins

import (
	"path/filepath"
	"plugin"

	"github.com/riotpot/pkg/logger"
	"github.com/riotpot/pkg/proxy"
	"github.com/riotpot/pkg/service"
)

var (
	pluginOffset = 20_000
)

// Function to get an stored service plugin.
// Note: the symbol used to get the plugin is "Name", which must be present in
// the plugin, and return type `Service` interface.
// based on: https://echorand.me/posts/getting-started-with-golang-plugins/
func getServicePlugin(path string) service.Service {

	// Open the plugin within the path
	pg, err := plugin.Open(path)
	logger.Log.Fatal().Err(err)

	// check the name of the function that exports the service
	// The plugin *Must* contain a variable called `Plugin`.
	s, err := pg.Lookup("Plugin")
	logger.Log.Fatal().Err(err)

	// log the name of the plugin being loaded
	logger.Log.Info().Msgf("Loading plugin: %s...\n", *s.(*string))

	// check if the reference symbol exists in the plugin
	rf, err := pg.Lookup(*s.(*string))
	logger.Log.Error().Err(err)

	// Load the service in a variable as the interface Service.
	newservice := rf.(func() service.Service)()

	// Assign the port of the plugin with an offset.
	// This will help creating the proxy while hiding the service.
	newservice.SetPort(newservice.GetPort() + pluginOffset)

	return newservice
}

// Get the plugin services included in the app
func GetPluginServices(pathLike string) (services []service.Service, err error) {
	// Get the paths to the plugins
	paths, err := filepath.Glob(pathLike)
	if err != nil {
		return
	}

	// Get the actual plugin and add it to the slice
	for _, path := range paths {
		service := getServicePlugin(path)
		services = append(services, service)
	}

	return
}

func LoadPlugins(pluginPath string) (proxies []proxy.Proxy, err error) {
	// Discover the services available to riotpot (running and stopped)
	plugins, err := GetPluginServices(pluginPath)
	if err != nil {
		return nil, err
	}

	// Add/register the plugin services
	plugins, err = service.Services.AddServices(plugins...)
	if err != nil {
		return nil, err
	}

	// Start the plugins
	for _, plugin := range plugins {
		_, errors := service.Services.Start(plugin.GetID())
		for _, err := range errors {
			logger.Log.Error().Err(err)
			return nil, err
		}
	}

	// Create proxies for each of the started plugins
	for _, service := range plugins {
		px, err := proxy.Proxies.CreateProxy(service.GetNetwork(), service.GetPort()-pluginOffset)
		if err != nil {
			return nil, err
		}

		// Add the service to the proxy
		px.SetService(service)
		proxies = append(proxies, px)
	}

	return
}
