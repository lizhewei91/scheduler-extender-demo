package scheduler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/klog/v2"

	"scheduler-extender-demo/common"
	"scheduler-extender-demo/pkg/route"
	"scheduler-extender-demo/pkg/scheduler/extender"
	"scheduler-extender-demo/pkg/storage"
)

type Scheduler struct {
	name       string
	router     *httprouter.Router
	filter     *extender.Filter
	prioritize *extender.Prioritize
}

func NewScheduler(storage *storage.Storage) *Scheduler {
	s := &Scheduler{
		name:       common.SchedulerName,
		router:     httprouter.New(),
		filter:     extender.NewFilter(storage),
		prioritize: extender.NewPrioritize(storage),
	}

	s.router.GET(common.VersionPath, route.VersionRoute)
	s.router.POST(common.FilterPath, route.FilterRouter(*s.filter))
	s.router.POST(common.PrioritizePath, route.PrioritizeRouter(*s.prioritize))

	return s
}

func (s *Scheduler) Run(addr string) error {
	klog.V(1).Info("scheduler start")

	return http.ListenAndServe(addr, s.router)
}
