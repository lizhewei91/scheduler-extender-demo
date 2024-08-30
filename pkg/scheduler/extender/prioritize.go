package extender

import (
	"math/rand"

	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"scheduler-extender-demo/common"
	"scheduler-extender-demo/pkg/storage"
)

const (
	// lucky priority gives a random [0, extender.MaxPriority] score
	// currently extender.MaxPriority is 10
	luckyPrioMsg = "pod: %v/%v is lucky, get score: %v\n"
)

type Prioritize struct {
	Name    string
	storage *storage.Storage
}

func NewPrioritize(storage *storage.Storage) *Prioritize {
	return &Prioritize{
		Name:    common.PrioritizeName,
		storage: storage,
	}
}

func (f *Prioritize) Handler(args extenderv1.ExtenderArgs) *extenderv1.HostPriorityList {
	pod := args.Pod
	nodeNames := *args.NodeNames

	if pod == nil {
		klog.Errorf("the filed pod of ExtenderArgs can't be empty")
		return &extenderv1.HostPriorityList{}
	}

	hostPriorityList := make(extenderv1.HostPriorityList, len(nodeNames))
	for i, nodeName := range nodeNames {
		score := rand.Int63n(extenderv1.MaxExtenderPriority + 1)
		klog.V(0).Infof(luckyPrioMsg, pod.Name, pod.Namespace, score)
		hostPriorityList[i] = extenderv1.HostPriority{
			Host:  nodeName,
			Score: score,
		}
	}
	return &hostPriorityList
}
