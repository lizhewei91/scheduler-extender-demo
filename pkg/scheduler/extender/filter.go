package extender

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"scheduler-extender-demo/common"
	"scheduler-extender-demo/pkg/storage"
	"scheduler-extender-demo/pkg/utils/util"
)

type Filter struct {
	Name    string
	storage *storage.Storage
}

func NewFilter(storage *storage.Storage) *Filter {
	return &Filter{
		Name:    common.FilterName,
		storage: storage,
	}
}

func (f *Filter) Handler(args extenderv1.ExtenderArgs) *extenderv1.ExtenderFilterResult {
	if args.Pod == nil {
		return &extenderv1.ExtenderFilterResult{
			Error: fmt.Errorf("the filed pod of ExtenderArgs can't be empty").Error(),
		}
	}

	fulleKey, err := common.KeyFunc(args.Pod)
	if err != nil {
		return &extenderv1.ExtenderFilterResult{
			Error: fmt.Errorf("parse Obj to storage key failed").Error(),
		}
	}

	// pod annotations 中存在 nodeNames 的处理逻辑：
	nodeNamesOfStorage, exist := f.storage.GetPodOfStorage(fulleKey)
	if exist {
		klog.V(1).Infof("%s is exist of storage, can be filtered", fulleKey)
		klog.V(1).Infof("will check node exists: %s", nodeNamesOfStorage)

		if args.NodeNames != nil {
			// 找出 args.NodeNames 参数 与 pod 对应 nodeNames 中，相同的元素
			commonNodeNames := util.CommonElements(*args.NodeNames, nodeNamesOfStorage)
			// 更新缓存
			f.storage.AddOrUpdatePodOfStorage(fulleKey, commonNodeNames)
			return &extenderv1.ExtenderFilterResult{
				NodeNames: &commonNodeNames,
			}
		}

		if args.Nodes != nil {
			eleMap := make(map[string]bool)
			for _, name := range nodeNamesOfStorage {
				eleMap[name] = true
			}
			for _, node := range args.Nodes.Items {
				if _, ok := eleMap[node.Name]; ok {
					return &extenderv1.ExtenderFilterResult{
						Nodes: &corev1.NodeList{
							Items: []corev1.Node{
								node,
							}},
					}
				}
			}
		}

		return &extenderv1.ExtenderFilterResult{
			Error: fmt.Errorf("nodeName fetch failed").Error(),
		}
	}

	return &extenderv1.ExtenderFilterResult{
		Nodes:     args.Nodes,
		NodeNames: args.NodeNames,
	}
}
