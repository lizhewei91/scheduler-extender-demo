package route

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"scheduler-extender-demo/common"
	"scheduler-extender-demo/pkg/scheduler/extender"
)

func FilterRouter(filter extender.Filter) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		checkBody(w, r)

		bodyBates, _ := io.ReadAll(r.Body)
		klog.V(0).Infof("FilterName: %s, ExtenderArgs: %s", filter.Name, string(bodyBates))

		var extenderArgs extenderv1.ExtenderArgs
		var extenderFilterResult *extenderv1.ExtenderFilterResult

		if err := json.Unmarshal(bodyBates, &extenderArgs); err != nil {
			extenderFilterResult = &extenderv1.ExtenderFilterResult{
				Nodes:       nil,
				NodeNames:   nil,
				FailedNodes: nil,
				Error:       err.Error(),
			}
		} else {
			extenderFilterResult = filter.Handler(extenderArgs)
		}

		if response, err := json.Marshal(extenderFilterResult); err != nil {
			klog.Error(err)
		} else {
			klog.V(0).Infof("FilterName: %s, ExtenderFilterResult: %s", filter.Name, string(response))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}
	}
}

func PrioritizeRouter(prioritize extender.Prioritize) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		checkBody(w, r)

		bodyBates, _ := io.ReadAll(r.Body)
		klog.V(0).Infof("prioritizeName: %s, ExtenderArgs: %s", prioritize.Name, string(bodyBates))

		var extenderArgs extenderv1.ExtenderArgs
		var hostPriorityList *extenderv1.HostPriorityList

		if err := json.Unmarshal(bodyBates, &extenderArgs); err != nil {
			klog.Error(err)
			hostPriorityList = &extenderv1.HostPriorityList{}
		} else {
			hostPriorityList = prioritize.Handler(extenderArgs)
		}

		if response, err := json.Marshal(hostPriorityList); err != nil {
			klog.Error(err)
		} else {
			klog.V(0).Infof("prioritizeName: %s, ExtenderPrioritizeResult: %s", prioritize.Name, string(response))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}
	}
}

func VersionRoute(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, fmt.Sprint(common.Version))
}

func checkBody(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "http request body can't be empty", 400)
		return
	}
}
