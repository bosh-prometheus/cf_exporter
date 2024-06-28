package collectors

import (
	"github.com/cloudfoundry/cf_exporter/fetcher"
	"github.com/cloudfoundry/cf_exporter/filters"
	"github.com/cloudfoundry/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

type ObjectCollector interface {
	Collect(*models.CFObjects, chan<- prometheus.Metric)
	Describe(ch chan<- *prometheus.Desc)
}

type Collector struct {
	workers    int
	cfConfig   *fetcher.CFConfig
	bbsConfig  *fetcher.BBSConfig
	filter     *filters.Filter
	collectors []ObjectCollector
}

func NewCollector(
	namespace string,
	environment string,
	deployment string,
	workers int,
	cfConfig *fetcher.CFConfig,
	bbsConfig *fetcher.BBSConfig,
	filter *filters.Filter,
) (*Collector, error) {
	res := &Collector{
		workers:    workers,
		cfConfig:   cfConfig,
		bbsConfig:  bbsConfig,
		filter:     filter,
		collectors: []ObjectCollector{},
	}

	if filter.Enabled(filters.Applications) {
		collector := NewApplicationsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Buildpacks) {
		collector := NewBuildpacksCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Domains) {
		collector := NewDomainsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.IsolationSegments) {
		collector := NewIsolationSegmentsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Organizations) {
		collector := NewOrganizationsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Routes) {
		collector := NewRoutesCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.SecurityGroups) {
		collector := NewSecurityGroupsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.ServiceBindings) {
		collector := NewServiceBindingsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.ServiceRouteBindings) {
		collector := NewRouteBindingsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.ServiceInstances) {
		collector := NewServiceInstancesCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.ServicePlans) {
		collector := NewServicePlansCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Services) {
		collector := NewServicesCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Stacks) {
		collector := NewStacksCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Spaces) {
		collector := NewSpacesCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Tasks) {
		collector := NewTasksCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	if filter.Enabled(filters.Events) {
		collector := NewEventsCollector(namespace, environment, deployment)
		res.collectors = append(res.collectors, collector)
	}

	return res, nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	fetcher := fetcher.NewFetcher(c.workers, c.cfConfig, c.bbsConfig, c.filter)
	objs := fetcher.GetObjects()

	for _, collector := range c.collectors {
		collector.Collect(objs, ch)
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range c.collectors {
		collector.Describe(ch)
	}
}
