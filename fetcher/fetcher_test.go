package fetcher

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/cloudfoundry/cf_exporter/filters"
)

var _ = ginkgo.Describe("Fetcher", func() {
	ginkgo.Context("fetching jobs are planned according to filter", func() {
		var (
			fetcher  *Fetcher
			active   []string
			jobs     []string
			expected []string
		)

		ginkgo.JustBeforeEach(func() {
			f, err := filters.NewFilter(active...)
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			fetcher = NewFetcher(10, &CFConfig{}, &BBSConfig{}, f)
			gomega.Ω(fetcher).ShouldNot(gomega.BeNil())
			fetcher.workInit()

			close(fetcher.worker.list)
			jobs = []string{}
			for w := range fetcher.worker.list {
				jobs = append(jobs, w.name)
			}
		})

		ginkgo.When("default filters are set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{}
				expected = []string{
					"info",
					"organizations",
					"org_quotas",
					"spaces",
					"space_quotas",
					"applications",
					"domains",
					"process",
					"routes",
					"route_services",
					"security_groups",
					"stacks",
					"buildpacks",
					"service_brokers",
					"service_offerings",
					"service_instances",
					"service_plans",
					"service_bindings",
					"service_route_bindings",
					"segments",
				}
			})
			ginkgo.It("plans all jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("all filters are set", func() {
			ginkgo.BeforeEach(func() {
				active = filters.All
				expected = []string{
					"info",
					"organizations",
					"org_quotas",
					"spaces",
					"space_quotas",
					"applications",
					"domains",
					"process",
					"routes",
					"route_services",
					"security_groups",
					"stacks",
					"buildpacks",
					"tasks",
					"service_brokers",
					"service_offerings",
					"service_instances",
					"service_plans",
					"service_bindings",
					"service_route_bindings",
					"segments",
					"users",
					"events",
				}
			})
			ginkgo.It("plans all jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("org filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Organizations}
				expected = []string{"info", "organizations", "org_quotas"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("space filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Spaces}
				expected = []string{"info", "spaces", "space_quotas"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("buildpack filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Buildpacks}
				expected = []string{"info", "buildpacks"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("tasks filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Tasks}
				expected = []string{"info", "tasks"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("isolationsegments filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.IsolationSegments}
				expected = []string{"info", "segments"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("routes filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Routes}
				expected = []string{"info", "routes", "route_services"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("securitygroups filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.SecurityGroups}
				expected = []string{"info", "security_groups"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("servicebindings filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.ServiceBindings}
				expected = []string{"info", "service_bindings"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("serviceinstances filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.ServiceInstances}
				expected = []string{"info", "service_instances"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("services filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Services}
				expected = []string{"info", "service_brokers", "service_offerings"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("stacks filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Stacks}
				expected = []string{"info", "stacks"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("applications filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Applications}
				expected = []string{"info", "organizations", "spaces", "applications", "process"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

		ginkgo.When("events filter is set", func() {
			ginkgo.BeforeEach(func() {
				active = []string{filters.Events}
				expected = []string{"info", "users", "events"}
			})
			ginkgo.It("plans only specific jobs", func() {
				gomega.Ω(jobs).Should(gomega.ConsistOf(expected))
			})
		})

	})
})
