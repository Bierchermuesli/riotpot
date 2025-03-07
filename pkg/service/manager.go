/*
This package contains implementations of the services manager
*/
package service

import (
	"fmt"

	lr "github.com/riotpot/pkg/logger"
	"github.com/riotpot/pkg/utils"
	"golang.org/x/exp/slices"
)

var (
	Services = NewServiceManager()
)

func IsRemovableService(service Service) (isRemovable bool) {
	// Add here the interfaces of services that should not be removable
	switch service.(type) {
	// Whether the service has the interface of a plugin
	case PluginService:
		isRemovable = true
	}

	return
}

type ServiceManager interface {
	// Register services
	AddServices(services ...Service) (serv []Service, err error)

	CreateService(name string, port int, network utils.Network, host string, interaction utils.Interaction) (Service, error)

	// Delete a service
	DeleteService(id string) (err error)

	// Get the list of services by their name
	GetServices() []Service

	// Get the list of plugin IDs registered
	GetPluginIDs() []string

	// Get a single service
	GetService(id string) (Service, error)

	// Start the plugin services
	Start(ids ...string) ([]Service, []error)
}

type serviceManager struct {
	ServiceManager

	// Set of services registered
	services []Service
}

// Add a service to the services map if it did not exist
func (se *serviceManager) AddServices(services ...Service) (serv []Service, err error) {
	// Returns a list of ID strings
	getServicesIDs := func(services []Service) (servs []string) {
		for _, service := range services {
			servs = append(servs, service.GetID())
		}
		return
	}

	// Convert the registered services into a simple array
	registeredIDs := getServicesIDs(se.GetServices())

	// Iterate the slice of services provided and add them to the services map
	for _, service := range services {
		serviceID := service.GetID()

		// Check whether the service is registered, and if not, add it to the list
		if !slices.Contains(registeredIDs, serviceID) {
			serv = append(serv, service)
		}
	}

	// Add the services to the registered services array
	se.services = append(se.services, serv...)

	return
}

// Creates a new service and register it in the manager
func (se *serviceManager) CreateService(name string, port int, network utils.Network, host string, interaction utils.Interaction) (s Service, err error) {
	// Iterate the services to determine whether the
	for _, service := range se.GetServices() {
		// Validate the name and interaction
		if service.GetName() == name && service.GetInteraction().String() == interaction.String() {
			err = fmt.Errorf("service already included: %s (%s) - %s", name, interaction.String(), service.GetInteraction().String())
			return
		}

		// Validate the address
		if service.GetPort() == port && service.GetNetwork() == network && service.GetHost() == host {
			err = fmt.Errorf("service address already taken: %s:%d", network.String(), port)
			return
		}
	}

	// Create the new service
	s = NewService(name, port, network, host, interaction)
	_, err = se.AddServices(s)
	if err != nil {
		return nil, err
	}

	return
}

// Remove a service from the list of registered
func (se *serviceManager) DeleteService(id string) (err error) {
	// Get all the services
	services := se.GetServices()

	for ind, service := range services {

		// Check if the service id is equal to the one received
		if service.GetID() == id {

			if service.IsLocked() {
				// If it was not found by this point, return an error
				err = fmt.Errorf("service locked")
				return
			}

			// Remove it from the slice by replacing it with the last item from the slice,
			// and reducing the slice by 1 element
			lastInd := len(services) - 1

			services[ind] = services[lastInd]
			se.services = services[:lastInd]

			return
		}
	}

	// If it was not found by this point, return an error
	err = fmt.Errorf("service not found")
	return
}

// Get services by name from the list of registered services
func (se *serviceManager) GetServices() []Service {
	return se.services
}

func (se *serviceManager) GetService(id string) (ret Service, err error) {
	for _, service := range se.GetServices() {
		if service.GetID() == id {
			ret = service
			return
		}
	}

	// If it was not found by this point, return an error
	err = fmt.Errorf("service not found")
	return
}

// Start each of the given Plugin Services by ID.
// Returns both arrays of errors and the started services
func (se *serviceManager) Start(ids ...string) (servs []Service, err []error) {
	for _, id := range ids {
		serv, e := se.GetService(id)
		if e != nil {
			err = append(err, e)
		}

		i, ok := serv.(PluginService)
		// If the service is not a plugin return an error
		if !ok {
			err = append(err, fmt.Errorf("service %s can not be started", serv.GetName()))
		}

		// Run the service
		go i.Run()

		lr.Log.Log().Msg(fmt.Sprintf("Service %s started", serv.GetName()))
		servs = append(servs, serv)
	}

	return
}

// Create a new pointer to a supervisor
func NewServiceManager() (manager ServiceManager) {
	// Initialise the manager
	manager = &serviceManager{
		services: []Service{},
	}

	return
}
