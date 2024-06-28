package fetcher

import (
	"sync"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/cloudfoundry/cf_exporter/filters"
	"github.com/cloudfoundry/cf_exporter/models"
	log "github.com/sirupsen/logrus"
)

var (
	LargeQuery = ccv3.Query{
		Key:    ccv3.PerPage,
		Values: []string{"5000"},
	}
	SortDesc = ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{"-created_at"},
	}
	TaskActiveStates = ccv3.Query{
		Key:    ccv3.StatesFilter,
		Values: []string{"PENDING", "RUNNING", "CANCELING"},
	}
)

type CFConfig struct {
	SkipSSLValidation bool   `yaml:"skip_ssl_validation"`
	URL               string `yaml:"url"`
	ClientID          string `yaml:"client_id"`
	ClientSecret      string `yaml:"client_secret"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
}

type Fetcher struct {
	sync.Mutex
	cfConfig  *CFConfig
	bbsConfig *BBSConfig
	worker    *Worker
}

func NewFetcher(threads int, config *CFConfig, bbsConfig *BBSConfig, filter *filters.Filter) *Fetcher {
	return &Fetcher{
		cfConfig:  config,
		bbsConfig: bbsConfig,
		worker:    NewWorker(threads, filter),
	}
}

func (c *Fetcher) GetObjects() *models.CFObjects {
	log.Infof("collecting objects from cloud foundry API")
	start := time.Now()
	data := c.fetch()
	took := time.Since(start).Seconds()
	log.Infof("collecting objects from cloud foundry API (done, %.0f sec)", took)
	data.Took = took
	return data
}

func (c *Fetcher) workInit() {
	c.worker.Reset()
	c.worker.Push("info", c.fetchInfo)
	c.worker.PushIf("organizations", c.fetchOrgs, filters.Applications, filters.Organizations)
	c.worker.PushIf("org_quotas", c.fetchOrgQuotas, filters.Organizations)
	c.worker.PushIf("spaces", c.fetchSpaces, filters.Applications, filters.Spaces)
	c.worker.PushIf("space_quotas", c.fetchSpaceQuotas, filters.Spaces)
	c.worker.PushIf("applications", c.fetchApplications, filters.Applications)
	c.worker.PushIf("droplets", c.fetchDroplets, filters.Droplets)
	c.worker.PushIf("domains", c.fetchDomains, filters.Domains)
	c.worker.PushIf("process", c.fetchProcesses, filters.Applications)
	c.worker.PushIf("routes", c.fetchRoutes, filters.Routes)
	c.worker.PushIf("route_services", c.fetchRouteServices, filters.Routes)
	c.worker.PushIf("security_groups", c.fetchSecurityGroups, filters.SecurityGroups)
	c.worker.PushIf("stacks", c.fetchStacks, filters.Stacks)
	c.worker.PushIf("buildpacks", c.fetchBuildpacks, filters.Buildpacks)
	c.worker.PushIf("tasks", c.fetchTasks, filters.Tasks)
	c.worker.PushIf("service_brokers", c.fetchServiceBrokers, filters.Services)
	c.worker.PushIf("service_offerings", c.fetchServiceOfferings, filters.Services)
	c.worker.PushIf("service_instances", c.fetchServiceInstances, filters.ServiceInstances)
	c.worker.PushIf("service_plans", c.fetchServicePlans, filters.ServicePlans)
	c.worker.PushIf("segments", c.fetchIsolationSegments, filters.IsolationSegments)
	c.worker.PushIf("service_bindings", c.fetchServiceBindings, filters.ServiceBindings)
	c.worker.PushIf("service_route_bindings", c.fetchServiceRouteBindings, filters.ServiceRouteBindings)
	c.worker.PushIf("users", c.fetchUsers, filters.Events)
	c.worker.PushIf("events", c.fetchEvents, filters.Events)
	c.worker.PushIf("actual_lrps", c.fetchActualLRPs)
}

func (c *Fetcher) fetch() *models.CFObjects {
	result := models.NewCFObjects()

	session, err := NewSessionExt(c.cfConfig)
	if err != nil {
		log.WithError(err).Error("unable to initialize cloud foundry clients")
		result.Error = err
		return result
	}
	bbs, err := NewBBSClient(c.bbsConfig)
	if err != nil {
		log.WithError(err).Error("unable to initialize bbs client")
		result.Error = err
		return result
	}

	c.workInit()

	result.Error = c.worker.Do(session, bbs, result)
	return result
}
