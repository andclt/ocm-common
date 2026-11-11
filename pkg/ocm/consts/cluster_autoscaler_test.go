package consts

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func mustParseDuration(value string) time.Duration {
	d, err := time.ParseDuration(value)
	Expect(err).NotTo(HaveOccurred(), "invalid duration %q", value)
	return d
}

// The defaults must be values Clusters Service accepts for a ROSA HCP cluster autoscaler.
var _ = Describe("ROSA HCP cluster autoscaler defaults", func() {
	Context("Phase-1 fields", func() {
		It("sets max_nodes_total to 0 (no limit) or at least 2", func() {
			Expect(HCPAutoscalerDefaultMaxNodesTotal == 0 || HCPAutoscalerDefaultMaxNodesTotal >= 2).To(BeTrue())
		})

		It("sets max_pod_grace_period to 0 or at least 600 seconds", func() {
			Expect(HCPAutoscalerDefaultMaxPodGracePeriod == 0 || HCPAutoscalerDefaultMaxPodGracePeriod >= 600).To(BeTrue())
		})

		It("sets max_node_provision_time between 15m and 60m", func() {
			d := mustParseDuration(HCPAutoscalerDefaultMaxNodeProvisionTime)
			Expect(d).To(BeNumerically(">=", 15*time.Minute))
			Expect(d).To(BeNumerically("<=", 60*time.Minute))
		})

		It("sets pod_priority_threshold to at most 1000000000", func() {
			Expect(HCPAutoscalerDefaultPodPriorityThreshold).To(BeNumerically("<=", 1000000000))
		})
	})

	Context("scale-down durations", func() {
		DescribeTable("are whole seconds between 0s and 24h",
			func(value string) {
				d := mustParseDuration(value)
				Expect(d).To(BeNumerically(">=", 0))
				Expect(d).To(BeNumerically("<=", 24*time.Hour))
				Expect(d % time.Second).To(BeZero())
			},
			Entry("scale_down.unneeded_time", HCPAutoscalerDefaultScaleDownUnneededTime),
			Entry("scale_down.delay_after_add", HCPAutoscalerDefaultScaleDownDelayAfterAdd),
			Entry("scale_down.delay_after_delete", HCPAutoscalerDefaultScaleDownDelayAfterDelete),
			Entry("scale_down.delay_after_failure", HCPAutoscalerDefaultScaleDownDelayAfterFailure),
		)
	})

	Context("ratios", func() {
		DescribeTable("are between 0 and 1 with at most two decimals",
			func(value string) {
				Expect(value).To(MatchRegexp(`^(0(\.[0-9]{1,2})?|1(\.0{1,2})?)$`))
			},
			Entry("scale_down.utilization_threshold", HCPAutoscalerDefaultScaleDownUtilizationThreshold),
			Entry("max_free_difference_ratio", HCPAutoscalerDefaultMaxFreeDifferenceRatio),
		)
	})

	Context("expanders", func() {
		It("has 1 to 3 entries", func() {
			Expect(len(HCPAutoscalerDefaultExpanders())).To(BeNumerically(">=", 1))
			Expect(len(HCPAutoscalerDefaultExpanders())).To(BeNumerically("<=", 3))
		})

		It("only uses LeastWaste, Priority and Random", func() {
			for _, e := range HCPAutoscalerDefaultExpanders() {
				Expect(e).To(BeElementOf("LeastWaste", "Priority", "Random"))
			}
		})

		It("has no duplicated entries", func() {
			seen := map[string]bool{}
			for _, e := range HCPAutoscalerDefaultExpanders() {
				Expect(seen[e]).To(BeFalse(), "expander %q is duplicated", e)
				seen[e] = true
			}
		})

		It("has Random last when other expanders are present", func() {
			expanders := HCPAutoscalerDefaultExpanders()
			for i, e := range expanders {
				if e == "Random" {
					Expect(i).To(Equal(len(expanders)-1), "Random must be the last expander")
				}
			}
		})

		It("returns a new slice on every call", func() {
			expanders := HCPAutoscalerDefaultExpanders()
			expanders[0] = "Random"
			Expect(HCPAutoscalerDefaultExpanders()).To(Equal([]string{"Priority", "LeastWaste"}))
		})
	})

	Context("balancing ignored labels", func() {
		It("is an empty, non-nil list", func() {
			labels := HCPAutoscalerDefaultBalancingIgnoredLabels()
			Expect(labels).NotTo(BeNil())
			Expect(labels).To(BeEmpty())
		})
	})
})
