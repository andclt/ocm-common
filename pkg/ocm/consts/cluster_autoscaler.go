package consts

// Default values of the ROSA HCP cluster autoscaler, shared by OCM services and clients.
//
// Values use the OCM API representation (/api/clusters_mgmt/v1/clusters/{id}/autoscaler):
// durations are Go duration strings and ratios are decimal strings between "0.0" and "1.0".
const (
	// HCPAutoscalerDefaultMaxNodesTotal is the default maximum number of worker nodes across all
	// machine pools (resource_limits.max_nodes_total). 0 means no limit.
	HCPAutoscalerDefaultMaxNodesTotal = 0

	// HCPAutoscalerDefaultMaxPodGracePeriod is the default time, in seconds, given to pods to
	// terminate when a node is removed (max_pod_grace_period).
	HCPAutoscalerDefaultMaxPodGracePeriod = 600

	// HCPAutoscalerDefaultMaxNodeProvisionTime is the default time to wait for a new node before
	// the scale up is considered failed (max_node_provision_time).
	HCPAutoscalerDefaultMaxNodeProvisionTime = "15m"

	// HCPAutoscalerDefaultPodPriorityThreshold is the default priority below which pods don't
	// trigger a scale up and don't prevent a node from being removed (pod_priority_threshold).
	HCPAutoscalerDefaultPodPriorityThreshold = -10

	// HCPAutoscalerDefaultScaleDownEnabled reports whether scale down is enabled by default
	// (scale_down.enabled).
	HCPAutoscalerDefaultScaleDownEnabled = true

	// HCPAutoscalerDefaultScaleDownUnneededTime is the default time a node must be underused before
	// it can be removed (scale_down.unneeded_time).
	HCPAutoscalerDefaultScaleDownUnneededTime = "10m"

	// HCPAutoscalerDefaultScaleDownUtilizationThreshold is the default node utilization below which
	// a node can be removed (scale_down.utilization_threshold).
	HCPAutoscalerDefaultScaleDownUtilizationThreshold = "0.5"

	// HCPAutoscalerDefaultScaleDownDelayAfterAdd is the default time to wait after a scale up before
	// evaluating scale down (scale_down.delay_after_add).
	HCPAutoscalerDefaultScaleDownDelayAfterAdd = "10m"

	// HCPAutoscalerDefaultScaleDownDelayAfterDelete is the default time to wait after a node is
	// removed before the next scale down (scale_down.delay_after_delete).
	HCPAutoscalerDefaultScaleDownDelayAfterDelete = "0s"

	// HCPAutoscalerDefaultScaleDownDelayAfterFailure is the default time to wait after a failed
	// scale down before retrying (scale_down.delay_after_failure).
	HCPAutoscalerDefaultScaleDownDelayAfterFailure = "3m"

	// HCPAutoscalerDefaultMaxFreeDifferenceRatio is the default maximum difference in free resources
	// for two node groups to be considered similar (max_free_difference_ratio).
	HCPAutoscalerDefaultMaxFreeDifferenceRatio = "0.1"
)

// HCPAutoscalerDefaultExpanders returns the default ordered list of scale up strategies (expanders).
// It returns a new slice on every call, so callers can't change the default.
func HCPAutoscalerDefaultExpanders() []string {
	return []string{"Priority", "LeastWaste"}
}

// HCPAutoscalerDefaultBalancingIgnoredLabels returns the default list of labels ignored when comparing
// node groups (balancing_ignored_labels), which is empty. It returns a new slice on every call.
func HCPAutoscalerDefaultBalancingIgnoredLabels() []string {
	return []string{}
}
