package common

import (
	"k8s.io/client-go/tools/cache"
)

const (
	SchedulerName = "customScheduler"
	Version       = "v0.0.1"
	VersionPath   = "/version"

	APIPrefix = "custom-scheduler"

	FilterName = "customFilter"
	FilterPath = APIPrefix + "/filter"

	PrioritizeName = "customPrioritize"
	PrioritizePath = APIPrefix + "/Prioritize"
)

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)
