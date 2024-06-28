// collectors/applications.go

package collectors

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/cloudfoundry/cf_exporter/models"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type ApplicationsCollector struct {
	namespace                                   string
	environment                                 string
	deployment                                  string
	applicationInfoMetric                       *prometheus.GaugeVec
	applicationBuildpackMetric                  *prometheus.GaugeVec
	applicationInstancesMetric                  *prometheus.GaugeVec
	applicationInstancesRunningMetric           *prometheus.GaugeVec
	applicationMemoryMbMetric                   *prometheus.GaugeVec
	applicationDiskQuotaMbMetric                *prometheus.GaugeVec
	applicationsScrapesTotalMetric              prometheus.Counter
	applicationsScrapeErrorsTotalMetric         prometheus.Counter
	lastApplicationsScrapeErrorMetric           prometheus.Gauge
	lastApplicationsScrapeTimestampMetric       prometheus.Gauge
	lastApplicationsScrapeDurationSecondsMetric prometheus.Gauge
}

func NewApplicationsCollector(
	namespace string,
	environment string,
	deployment string,
) *ApplicationsCollector {
	applicationInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "application",
			Name:        "info",
			Help:        "Labeled Cloud Foundry Application information with a constant '1' value.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "application_name", "detected_buildpack", "buildpack", "organization_id", "organization_name", "space_id", "space_name", "stack_id", "state"},
	)

	applicationBuildpackMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "application",
			Name:        "buildpack",
			Help:        "Buildpack used by an Application.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "application_name", "buildpack_name", "detected_buildpack"},
	)

	applicationInstancesMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "application",
			Name:        "instances",
			Help:        "Number of desired Cloud Foundry Application Instances.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name", "state"},
	)

	applicationInstancesRunningMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "application",
			Name:        "instances_running",
			Help:        "Number of running Cloud Foundry Application Instances.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name", "state"},
	)

	applicationMemoryMbMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "application",
			Name:        "memory_mb",
			Help:        "Cloud Foundry Application Memory (Mb).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name"},
	)

	applicationDiskQuotaMbMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "application",
			Name:        "disk_quota_mb",
			Help:        "Cloud Foundry Application Disk Quota (Mb).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
		[]string{"application_id", "application_name", "organization_id", "organization_name", "space_id", "space_name"},
	)

	applicationsScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "applications_scrapes",
			Name:        "total",
			Help:        "Total number of scrapes for Cloud Foundry Applications.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	applicationsScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   "applications_scrape_errors",
			Name:        "total",
			Help:        "Total number of scrape errors of Cloud Foundry Applications.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastApplicationsScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_applications_scrape_error",
			Help:        "Whether the last scrape of Applications metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastApplicationsScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_applications_scrape_timestamp",
			Help:        "Number of seconds since 1970 since last scrape of Applications metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	lastApplicationsScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "",
			Name:        "last_applications_scrape_duration_seconds",
			Help:        "Duration of the last scrape of Applications metrics from Cloud Foundry.",
			ConstLabels: prometheus.Labels{"environment": environment, "deployment": deployment},
		},
	)

	return &ApplicationsCollector{
		namespace:                                   namespace,
		environment:                                 environment,
		deployment:                                  deployment,
		applicationInfoMetric:                       applicationInfoMetric,
		applicationBuildpackMetric:                  applicationBuildpackMetric,
		applicationInstancesMetric:                  applicationInstancesMetric,
		applicationInstancesRunningMetric:           applicationInstancesRunningMetric,
		applicationMemoryMbMetric:                   applicationMemoryMbMetric,
		applicationDiskQuotaMbMetric:                applicationDiskQuotaMbMetric,
		applicationsScrapesTotalMetric:              applicationsScrapesTotalMetric,
		applicationsScrapeErrorsTotalMetric:         applicationsScrapeErrorsTotalMetric,
		lastApplicationsScrapeErrorMetric:           lastApplicationsScrapeErrorMetric,
		lastApplicationsScrapeTimestampMetric:       lastApplicationsScrapeTimestampMetric,
		lastApplicationsScrapeDurationSecondsMetric: lastApplicationsScrapeDurationSecondsMetric,
	}
}

func (c ApplicationsCollector) Collect(objs *models.CFObjects, ch chan<- prometheus.Metric) {
	errorMetric := float64(0)
	if objs.Error != nil {
		errorMetric = float64(1)
		c.applicationsScrapeErrorsTotalMetric.Inc()
	} else {
		err := c.reportApplicationsMetrics(objs, ch)
		if err != nil {
			errorMetric = float64(1)
			c.applicationsScrapeErrorsTotalMetric.Inc()
		}
	}

	c.applicationsScrapeErrorsTotalMetric.Collect(ch)
	c.applicationsScrapesTotalMetric.Inc()
	c.applicationsScrapesTotalMetric.Collect(ch)
	c.lastApplicationsScrapeErrorMetric.Set(errorMetric)
	c.lastApplicationsScrapeErrorMetric.Collect(ch)
	c.lastApplicationsScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastApplicationsScrapeTimestampMetric.Collect(ch)
	c.lastApplicationsScrapeDurationSecondsMetric.Set(objs.Took)
	c.lastApplicationsScrapeDurationSecondsMetric.Collect(ch)
}

func (c ApplicationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.applicationInfoMetric.Describe(ch)
	c.applicationInstancesMetric.Describe(ch)
	c.applicationInstancesRunningMetric.Describe(ch)
	c.applicationMemoryMbMetric.Describe(ch)
	c.applicationDiskQuotaMbMetric.Describe(ch)
	c.applicationsScrapesTotalMetric.Describe(ch)
	c.applicationsScrapeErrorsTotalMetric.Describe(ch)
	c.applicationBuildpackMetric.Describe(ch)
	c.lastApplicationsScrapeErrorMetric.Describe(ch)
	c.lastApplicationsScrapeTimestampMetric.Describe(ch)
	c.lastApplicationsScrapeDurationSecondsMetric.Describe(ch)
}

// reportApplicationsMetrics
//  1. empty detected buildpacks for apps running on droplet
//     staged with a buildpack that is no mot available
//     fallback to buildpack field for compatibility with v0
//  2. symmetrically in some corner cases, buildpack is null but
//     detected_buildpack is available. Use detected_buildpack
//     for compatibility with v0
func (c ApplicationsCollector) reportApp(application models.Application, objs *models.CFObjects) error {
	processes, ok := objs.AppProcesses[application.GUID]
	if !ok {
		return fmt.Errorf("could not find processes for application '%s'", application.GUID)
	}
	process := processes[0]
	for _, cProc := range processes {
		if cProc.Type == "web" {
			process = cProc
		}
	}

	spaceRel, ok := application.Relationships[constant.RelationshipTypeSpace]
	if !ok {
		return fmt.Errorf("could not find space relation in application '%s'", application.GUID)
	}
	space, ok := objs.Spaces[spaceRel.GUID]
	if !ok {
		return fmt.Errorf("could not find space with guid '%s'", spaceRel.GUID)
	}
	orgRel, ok := space.Relationships[constant.RelationshipTypeOrganization]
	if !ok {
		return fmt.Errorf("could not find org relation in space '%s'", space.GUID)
	}
	organization, ok := objs.Orgs[orgRel.GUID]
	if !ok {
		return fmt.Errorf("could not find org with guid '%s'", orgRel.GUID)
	}

	detectedBuildpack := ""
	buildpack := ""
	stackGUID := ""
	for _, stack := range objs.Stacks {
		if stack.Name == application.Lifecycle.Data.Stack {
			stackGUID = stack.GUID
			break
		}
	}
	if dropletGUID := application.Relationships[constant.RelationshipTypeCurrentDroplet].GUID; dropletGUID != "" {
		if droplet, ok := objs.Droplets[dropletGUID]; ok {
			// 1.
			detectedBuildpack = droplet.Buildpacks[0].DetectOutput
			// 2.
			buildpack = droplet.Buildpacks[0].BuildpackName
			if len(detectedBuildpack) == 0 {
				detectedBuildpack = buildpack
			}
			if len(buildpack) == 0 {
				buildpack = detectedBuildpack
			}
			// 3.Use the droplet data for the buildpack metric
			for _, bp := range droplet.Buildpacks {
				c.applicationBuildpackMetric.WithLabelValues(
					application.GUID,
					application.Name,
					bp.BuildpackName,
					bp.DetectOutput,
				).Set(float64(1))
			}
		}
	}

	c.applicationInfoMetric.WithLabelValues(
		application.GUID,
		application.Name,
		detectedBuildpack,
		buildpack,
		organization.GUID,
		organization.Name,
		space.GUID,
		space.Name,
		stackGUID,
		string(application.State),
	).Set(float64(1))

	c.applicationInstancesMetric.WithLabelValues(
		application.GUID,
		application.Name,
		organization.GUID,
		organization.Name,
		space.GUID,
		space.Name,
		string(application.State),
	).Set(float64(process.Instances.Value))

	// Use bbs data if available
	runningInstances := 0
	if len(objs.ProcessActualLRPs) > 0 {
		LRPs, ok := objs.ProcessActualLRPs[process.GUID]
		if ok {
			for _, lrp := range LRPs {
				if lrp.State == "RUNNING" {
					runningInstances++
				}
			}
		}
	}
	c.applicationInstancesRunningMetric.WithLabelValues(
		application.GUID,
		application.Name,
		organization.GUID,
		organization.Name,
		space.GUID,
		space.Name,
		string(application.State),
	).Set(float64(runningInstances))

	c.applicationMemoryMbMetric.WithLabelValues(
		application.GUID,
		application.Name,
		organization.GUID,
		organization.Name,
		space.GUID,
		space.Name,
	).Set(float64(process.MemoryInMB.Value))

	c.applicationDiskQuotaMbMetric.WithLabelValues(
		application.GUID,
		application.Name,
		organization.GUID,
		organization.Name,
		space.GUID,
		space.Name,
	).Set(float64(process.DiskInMB.Value))
	return nil
}

// reportApplicationsMetrics
//  1. continue processing application list upon error
func (c ApplicationsCollector) reportApplicationsMetrics(objs *models.CFObjects, ch chan<- prometheus.Metric) error {
	var res error

	c.applicationInfoMetric.Reset()
	c.applicationInstancesMetric.Reset()
	c.applicationInstancesRunningMetric.Reset()
	c.applicationMemoryMbMetric.Reset()
	c.applicationDiskQuotaMbMetric.Reset()
	c.applicationBuildpackMetric.Reset()

	for _, application := range objs.Apps {
		err := c.reportApp(application, objs)
		// 1.
		if err != nil {
			log.Warn(err)
			res = err
		}
	}

	c.applicationInfoMetric.Collect(ch)
	c.applicationInstancesMetric.Collect(ch)
	c.applicationInstancesRunningMetric.Collect(ch)
	c.applicationMemoryMbMetric.Collect(ch)
	c.applicationDiskQuotaMbMetric.Collect(ch)
	c.applicationBuildpackMetric.Collect(ch)
	return res
}
