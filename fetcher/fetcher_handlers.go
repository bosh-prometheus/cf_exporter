package fetcher

import (
	"regexp"
	"time"

	models2 "code.cloudfoundry.org/bbs/models"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"github.com/cloudfoundry/cf_exporter/models"
	log "github.com/sirupsen/logrus"
)

func loadIndex[T any](store map[string]T, objects []T, key func(T) string) {
	for _, val := range objects {
		store[key(val)] = val
	}
}

func (c *Fetcher) fetchActualLRPs(_ *SessionExt, bbs *BBSClient, entry *models.CFObjects) error {
	if bbs == nil {
		return nil
	}
	log.Infof("fetching resources from BBS API")
	actualLRPs, err := bbs.GetActualLRPs()
	if err == nil {
		// match first guid as lrps process_guid field contains process_guid and instance_guid "<:process_guid>-<:instance_guid>"
		re := regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}")
		for idx := 0; idx < len(actualLRPs); idx++ {
			processGUID := actualLRPs[idx].ProcessGuid
			match := re.FindString(processGUID)
			if match != "" {
				processGUID = match
			}
			_, ok := entry.ProcessActualLRPs[processGUID]
			if !ok {
				entry.ProcessActualLRPs[processGUID] = []*models2.ActualLRP{}
			}
			entry.ProcessActualLRPs[processGUID] = append(entry.ProcessActualLRPs[processGUID], actualLRPs[idx])
		}
	} else {
		log.Errorf("could not fetch actual lrps: %s", err)
	}
	return err
}

func (c *Fetcher) fetchInfo(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	var err error
	entry.Info, err = session.GetInfo()
	return err
}

func (c *Fetcher) fetchOrgs(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	orgs, _, err := session.V3().GetOrganizations(LargeQuery)
	if err == nil {
		loadIndex(entry.Orgs, orgs, func(r resources.Organization) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchOrgQuotas(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	quotas, err := session.GetOrganizationQuotas()
	if err == nil {
		loadIndex(entry.OrgQuotas, quotas, func(r models.Quota) string { return r.GUID })
	}
	return err
}

// fetchSpaces
//  1. silent fail because space may have been deleted between listing and
//     summary fetching attempt. See cloudfoundry/cf_exporter#85
func (c *Fetcher) fetchSpaces(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	spaces, _, _, err := session.V3().GetSpaces(LargeQuery)
	if err == nil {
		loadIndex(entry.Spaces, spaces, func(r resources.Space) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchSpaceQuotas(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	quotas, err := session.GetSpaceQuotas()
	if err == nil {
		loadIndex(entry.SpaceQuotas, quotas, func(r models.Quota) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchApplications(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	apps, err := session.GetApplications()
	if err == nil {
		loadIndex(entry.Apps, apps, func(r models.Application) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchDomains(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	domains, _, err := session.V3().GetDomains(LargeQuery)
	if err == nil {
		loadIndex(entry.Domains, domains, func(r resources.Domain) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchProcesses(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	processes, _, err := session.V3().GetProcesses(LargeQuery)
	if err != nil {
		return err
	}

	loadIndex(entry.Processes, processes, func(r resources.Process) string { return r.GUID })
	for idx := 0; idx < len(processes); idx++ {
		appGUID := processes[idx].AppGUID
		_, ok := entry.AppProcesses[appGUID]
		if !ok {
			entry.AppProcesses[appGUID] = []resources.Process{}
		}
		entry.AppProcesses[appGUID] = append(entry.AppProcesses[appGUID], processes[idx])
	}
	return nil
}

func (c *Fetcher) fetchRoutes(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	routes, _, err := session.V3().GetRoutes(LargeQuery)
	if err == nil {
		loadIndex(entry.Routes, routes, func(r resources.Route) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchRouteServices(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	routes, _, _, err := session.V3().GetRouteBindings(LargeQuery)
	if err == nil {
		loadIndex(entry.RoutesBindings, routes, func(r resources.RouteBinding) string { return r.RouteGUID })
	}
	return err
}

func (c *Fetcher) fetchSecurityGroups(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	securitygroups, _, err := session.V3().GetSecurityGroups(LargeQuery)
	if err == nil {
		loadIndex(entry.SecurityGroups, securitygroups, func(r resources.SecurityGroup) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchDroplets(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	droplets, _, err := session.V3().GetDroplets(LargeQuery)
	if err == nil {
		loadIndex(entry.Droplets, droplets, func(r resources.Droplet) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchStacks(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	stacks, _, err := session.V3().GetStacks(LargeQuery)
	if err == nil {
		loadIndex(entry.Stacks, stacks, func(r resources.Stack) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchBuildpacks(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	buildpacks, _, err := session.V3().GetBuildpacks(LargeQuery)
	if err == nil {
		loadIndex(entry.Buildpacks, buildpacks, func(r resources.Buildpack) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchTasks(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	tasks, err := session.GetTasks()
	if err == nil {
		loadIndex(entry.Tasks, tasks, func(r models.Task) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchServiceBrokers(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	servicebrokers, _, err := session.V3().GetServiceBrokers(LargeQuery)
	if err == nil {
		loadIndex(entry.ServiceBrokers, servicebrokers, func(r resources.ServiceBroker) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchServiceOfferings(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	serviceofferings, _, err := session.V3().GetServiceOfferings(LargeQuery)
	if err == nil {
		loadIndex(entry.ServiceOfferings, serviceofferings, func(r resources.ServiceOffering) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchServiceInstances(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	serviceinstances, _, _, err := session.V3().GetServiceInstances(LargeQuery)
	if err == nil {
		loadIndex(entry.ServiceInstances, serviceinstances, func(r resources.ServiceInstance) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchServicePlans(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	plans, _, err := session.V3().GetServicePlans()
	if err == nil {
		loadIndex(entry.ServicePlans, plans, func(r resources.ServicePlan) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchServiceBindings(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	bindings, _, err := session.V3().GetServiceCredentialBindings(LargeQuery)
	if err == nil {
		loadIndex(entry.ServiceBindings, bindings, func(r resources.ServiceCredentialBinding) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchServiceRouteBindings(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	routeBindings, _, _, err := session.V3().GetRouteBindings(LargeQuery)
	if err == nil {
		loadIndex(entry.ServiceRouteBindings, routeBindings, func(r resources.RouteBinding) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchIsolationSegments(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	segments, _, err := session.V3().GetIsolationSegments()
	if err == nil {
		loadIndex(entry.Segments, segments, func(r resources.IsolationSegment) string { return r.GUID })
	}
	return err
}

func (c *Fetcher) fetchUsers(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	users, _, err := session.V3().GetUsers(LargeQuery)
	if err == nil {
		loadIndex(entry.Users, users, func(r resources.User) string { return r.GUID })
	}
	return err
}

// fetchEvents -
//  1. create query param "created_ats[gt]=(now - 15min)". There is no point scrapping more
//     data since the event metric will filter out events older than last scrap.
func (c *Fetcher) fetchEvents(session *SessionExt, _ *BBSClient, entry *models.CFObjects) error {
	// 1.
	location, _ := time.LoadLocation("UTC")
	since := time.Now().Add(-1 * 15 * time.Minute)
	newTime := since.In(location).Format("2006-01-02T15:04:05Z")
	recent := ccv3.Query{
		Key:    "created_ats[gt]",
		Values: []string{newTime},
	}

	events, err := session.GetEvents(LargeQuery, SortDesc, recent)
	if err == nil {
		loadIndex(entry.Events, events, func(r models.Event) string { return r.GUID })
	}
	return err
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
