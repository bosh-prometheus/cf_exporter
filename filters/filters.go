package filters

import (
	"fmt"
	"strings"
)

const (
	Applications         = "applications"
	Droplets             = "droplets"
	Buildpacks           = "buildpacks"
	Domains              = "domains"
	Events               = "events"
	IsolationSegments    = "isolationsegments"
	Organizations        = "organizations"
	Routes               = "routes"
	SecurityGroups       = "securitygroups"
	ServiceBindings      = "servicebindings"
	ServiceRouteBindings = "service_route_bindings"
	ServiceInstances     = "serviceinstances"
	ServicePlans         = "serviceplans"
	Services             = "services"
	Spaces               = "spaces"
	Stacks               = "stacks"
	Tasks                = "tasks"
)

var (
	All = []string{
		Applications,
		Droplets,
		Buildpacks,
		Domains,
		Events,
		IsolationSegments,
		Organizations,
		Routes,
		SecurityGroups,
		ServiceBindings,
		ServiceRouteBindings,
		ServiceInstances,
		ServicePlans,
		Services,
		Spaces,
		Stacks,
		Tasks,
	}
)

type Filter struct {
	activated map[string]bool
}

func NewFilter(active ...string) (*Filter, error) {
	filter := &Filter{
		activated: map[string]bool{
			Applications:         true,
			Droplets:             true,
			Buildpacks:           true,
			Domains:              true,
			IsolationSegments:    true,
			Organizations:        true,
			Routes:               true,
			SecurityGroups:       true,
			ServiceBindings:      true,
			ServiceRouteBindings: true,
			ServiceInstances:     true,
			ServicePlans:         true,
			Services:             true,
			Spaces:               true,
			Stacks:               true,
			Tasks:                false,
			Events:               false,
		},
	}

	if len(active) != 0 {
		if err := filter.setActive(active); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

func (f *Filter) setActive(active []string) error {
	// override default states with all disabled
	f.activated = map[string]bool{
		Applications:         false,
		Droplets:             false,
		Buildpacks:           false,
		Domains:              false,
		IsolationSegments:    false,
		Organizations:        false,
		Routes:               false,
		SecurityGroups:       false,
		ServiceBindings:      false,
		ServiceRouteBindings: false,
		ServiceInstances:     false,
		ServicePlans:         false,
		Services:             false,
		Spaces:               false,
		Stacks:               false,
		Tasks:                false,
		Events:               false,
	}

	// enable only given filters
	for _, val := range active {
		name := strings.Trim(val, " ")
		name = strings.ToLower(name)
		if _, ok := f.activated[name]; !ok {
			return fmt.Errorf("Filter `%s` is not supported", val)
		}
		f.activated[name] = true
	}
	return nil
}

func (f *Filter) Enabled(name string) bool {
	status, ok := f.activated[name]
	return ok && status
}

func (f *Filter) Any(names ...string) bool {
	for _, n := range names {
		if f.Enabled(n) {
			return true
		}
	}
	return false
}

func (f *Filter) All(names ...string) bool {
	for _, n := range names {
		if !f.Enabled(n) {
			return false
		}
	}
	return true
}
