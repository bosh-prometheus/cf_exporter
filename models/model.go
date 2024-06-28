package models

import (
	"time"

	"code.cloudfoundry.org/bbs/models"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type CFObjects struct {
	Info                 Info                                          `json:"info"`
	Orgs                 map[string]resources.Organization             `json:"orgs"`
	OrgQuotas            map[string]Quota                              `json:"org_quotas"`
	Spaces               map[string]resources.Space                    `json:"spaces"`
	SpaceQuotas          map[string]Quota                              `json:"space_quotas"`
	Apps                 map[string]Application                        `json:"apps"`
	Droplets             map[string]resources.Droplet                  `json:"droplets"`
	Processes            map[string]resources.Process                  `json:"process"`
	Tasks                map[string]Task                               `json:"tasks"`
	Routes               map[string]resources.Route                    `json:"routes"`
	RoutesBindings       map[string]resources.RouteBinding             `json:"route_bindings"`
	Segments             map[string]resources.IsolationSegment         `json:"segments"`
	ServiceInstances     map[string]resources.ServiceInstance          `json:"service_instances"`
	SecurityGroups       map[string]resources.SecurityGroup            `json:"security_groups"`
	Stacks               map[string]resources.Stack                    `json:"stacks"`
	Buildpacks           map[string]resources.Buildpack                `json:"buildpacks"`
	BuildpacksByName     map[string]resources.Buildpack                `json:"builpacks_by_name"`
	Domains              map[string]resources.Domain                   `json:"domains"`
	ServiceBrokers       map[string]resources.ServiceBroker            `json:"service_brokers"`
	ServiceOfferings     map[string]resources.ServiceOffering          `json:"service_offerings"`
	ServicePlans         map[string]resources.ServicePlan              `json:"service_plans"`
	ServiceBindings      map[string]resources.ServiceCredentialBinding `json:"service_bindings"`
	AppProcesses         map[string][]resources.Process                `json:"app_processes"`
	ProcessActualLRPs    map[string][]*models.ActualLRP                `json:"process_actual_lrps"`
	Events               map[string]Event                              `json:"events"`
	Users                map[string]resources.User                     `json:"users"`
	ServiceRouteBindings map[string]resources.RouteBinding             `json:"service_route_bindings"`
	Took                 float64
	Error                error
}

type QuotaApp struct {
	TotalMemory       *types.NullInt `json:"total_memory_in_mb,omitempty"`
	InstanceMemory    *types.NullInt `json:"per_process_memory_in_mb,omitempty"`
	TotalAppInstances *types.NullInt `json:"total_instances,omitempty"`
	PerAppTasks       *types.NullInt `json:"per_app_tasks,omitempty"`
}

type QuotaService struct {
	TotalServiceInstances *types.NullInt `json:"total_service_instances,omitempty"`
	TotalServiceKeys      *types.NullInt `json:"total_service_keys,omitempty"`
	PaidServicePlans      *bool          `json:"paid_services_allowed,omitempty"`
}

type QuotaDomain struct {
	TotalDomains *types.NullInt `json:"total_domains,omitempty"`
}

type Quota struct {
	GUID     string               `json:"guid,omitempty"`
	Name     string               `json:"name"`
	Apps     QuotaApp             `json:"apps"`
	Services QuotaService         `json:"services"`
	Routes   resources.RouteLimit `json:"routes"`
	Domains  QuotaDomain          `json:"domains"`
}

type Info struct {
	Name string `json:"name"`
}

type Metadata struct {
	Labels      map[string]types.NullString `json:"labels,omitempty"`
	Annotations map[string]types.NullString `json:"annotations,omitempty"`
}

type Lifecycle struct {
	Type constant.AppLifecycleType `json:"type,omitempty"`
	Data struct {
		Buildpacks []string `json:"buildpacks,omitempty"`
		Stack      string   `json:"stack,omitempty"`
	} `json:"data"`
}

type Application struct {
	GUID          string                    `json:"guid,omitempty"`
	Name          string                    `json:"name,omitempty"`
	State         constant.ApplicationState `json:"state,omitempty"`
	Metadata      *Metadata                 `json:"metadata,omitempty"`
	Relationships resources.Relationships   `json:"relationships,omitempty"`
	Lifecycle     Lifecycle                 `json:"lifecycle,omitempty"`
	CreatedAt     string                    `json:"created_at,omitempty"`
	UpdatedAt     string                    `json:"updated_at,omitempty"`
}

type Task struct {
	GUID          string                  `json:"guid,omitempty"`
	State         constant.TaskState      `json:"state,omitempty"`
	Relationships resources.Relationships `json:"relationships,omitempty"`
	CreatedAt     time.Time               `json:"created_at,omitempty"`
	MemoryInMb    int64                   `json:"memory_in_mb,omitempty"`
	DiskInMb      int64                   `json:"disk_in_mb,omitempty"`
}

type SpaceSummary struct {
	GUID string       `json:"guid,omitempty"`
	Name string       `json:"name,omitempty"`
	Apps []AppSummary `json:"apps,omitempty"`
}

type AppSummary struct {
	GUID              string `json:"guid,omitempty"`
	RunningInstances  int    `json:"running_instances,omitempty"`
	DetectedBuildpack string `json:"detected_buildpack,omitempty"`
	Buildpack         string `json:"buildpack,omitempty"`
	StackID           string `json:"stack_guid,omitempty"`
}

type EventActor struct {
	GUID string `json:"guid,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type EventTarget struct {
	GUID string `json:"guid,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type EventSpace struct {
	GUID string `json:"guid,omitempty"`
}

type EventOrg struct {
	GUID string `json:"guid,omitempty"`
}

type Event struct {
	GUID      string                 `json:"guid,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Actor     EventActor             `json:"actor,omitempty"`
	Target    EventTarget            `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Space     EventSpace             `json:"space,omitempty"`
	Org       EventOrg               `json:"organization,omitempty"`
}

func NewCFObjects() *CFObjects {
	return &CFObjects{
		Info:                 Info{},
		Orgs:                 map[string]resources.Organization{},
		OrgQuotas:            map[string]Quota{},
		Spaces:               map[string]resources.Space{},
		SpaceQuotas:          map[string]Quota{},
		Apps:                 map[string]Application{},
		Droplets:             map[string]resources.Droplet{},
		Processes:            map[string]resources.Process{},
		Tasks:                map[string]Task{},
		Routes:               map[string]resources.Route{},
		RoutesBindings:       map[string]resources.RouteBinding{},
		Segments:             map[string]resources.IsolationSegment{},
		ServiceInstances:     map[string]resources.ServiceInstance{},
		SecurityGroups:       map[string]resources.SecurityGroup{},
		Stacks:               map[string]resources.Stack{},
		Buildpacks:           map[string]resources.Buildpack{},
		Domains:              map[string]resources.Domain{},
		ServiceBrokers:       map[string]resources.ServiceBroker{},
		ServiceOfferings:     map[string]resources.ServiceOffering{},
		ServicePlans:         map[string]resources.ServicePlan{},
		ServiceBindings:      map[string]resources.ServiceCredentialBinding{},
		AppProcesses:         map[string][]resources.Process{},
		ProcessActualLRPs:    map[string][]*models.ActualLRP{},
		Users:                map[string]resources.User{},
		Events:               map[string]Event{},
		ServiceRouteBindings: map[string]resources.RouteBinding{},
		Took:                 0,
		Error:                nil,
	}
}
