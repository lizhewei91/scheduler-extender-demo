package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"scheduler-extender-demo/common"
	"scheduler-extender-demo/pkg/storage"
)

const (
	maxReQueue = 10
)

type Controller struct {
	client kubernetes.Interface

	podLister  corelister.PodLister
	nodeLister corelister.NodeLister

	podInformerSynced  cache.InformerSynced
	nodeInformerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	Storage *storage.Storage
}

func NewController(client kubernetes.Interface, podInformer coreinformer.PodInformer, nodeInformer coreinformer.NodeInformer) (*Controller, error) {
	c := &Controller{
		client:  client,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "scheduler controller"),
		Storage: storage.NewStorage(),
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addPod,
		UpdateFunc: c.updatePod,
		DeleteFunc: c.deletePod,
	})

	c.podLister = podInformer.Lister()
	c.nodeLister = nodeInformer.Lister()
	c.podInformerSynced = podInformer.Informer().HasSynced
	c.nodeInformerSynced = nodeInformer.Informer().HasSynced

	return c, nil
}

func (c *Controller) Run(ctx context.Context) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.V(1).Info("controller start")

	if !cache.WaitForCacheSync(ctx.Done(), c.podInformerSynced) {
		klog.Fatalf("podInformer isn't synced")
	}

	if !cache.WaitForCacheSync(ctx.Done(), c.nodeInformerSynced) {
		klog.Fatalf("nodeInformer isn't synced")
	}

	for i := 0; i < 5; i++ {
		go wait.Until(c.worker, time.Second, ctx.Done())
	}

	<-ctx.Done()
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Done(item)

	err := c.syncPod(item.(string))
	if err != nil {
		c.handleErr(item, err)
	}

	return true
}

func (c *Controller) addPod(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.V(1).Infof("can't convert %v to *v1.Pod", obj)
		return
	}

	klog.V(1).Infof("start add pod to queue: %s", pod.Name)
	c.enqueue(pod)
}

func (c *Controller) updatePod(oldObj interface{}, newObj interface{}) {
	oldPod, ok := oldObj.(*corev1.Pod)
	if !ok {
		klog.V(1).Infof("can't convert %v to *v1.Pod", oldObj)
		return
	}

	curPod, ok := newObj.(*corev1.Pod)
	if !ok {
		klog.V(1).Infof("can't convert %v to *v1.Pod", newObj)
		return
	}

	klog.V(1).Infof("start update pod to queue: %s", oldPod.Name)
	c.enqueue(curPod)
}

func (c *Controller) deletePod(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.V(1).Infof("can't convert %v to *v1.Pod", obj)
		return
	}

	klog.V(1).Infof("start delete pod to queue: %s", pod.Name)
	c.enqueue(pod)
}

func (c *Controller) enqueue(pod *corev1.Pod) {
	key, err := common.KeyFunc(pod)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", pod, err))
		return
	}

	klog.V(1).Infof("add %s to queue", key)
	c.queue.Add(key)
}

func (c *Controller) handleErr(item interface{}, err error) {
	if err == nil {
		return
	}

	if c.queue.NumRequeues(item) <= maxReQueue {
		c.queue.AddRateLimited(item)
		return
	}

	c.queue.Forget(item)
}

func (c *Controller) syncPod(key string) error {
	if err := c.Storage.DeletePodOfStorage(key); err != nil {
		return err
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.V(1).Infof("split key to namespace and name failed: %s", key)
		return err
	}

	startTime := time.Now()
	klog.V(1).Infof("start syncing pod: %s, start time: %v", name, startTime)
	defer func() {
		klog.V(1).Infof("finish syncing pod: %s, duration time: %v", name, time.Since(startTime))
	}()

	pod, err := c.podLister.Pods(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(1).Infof("pod is deleted: %s", name)
			return nil
		}
		klog.V(1).Infof("get pod failed: %s", name)
		return err
	}

	nodeNames, exist := pod.Annotations["nodeNames"]
	if !exist {
		if err := c.Storage.DeletePodOfStorage(key); err != nil {
			return err
		}

		klog.V(1).Infof("this pod needn't been synced: %s", name)
		return nil
	}

	nodeNameSlice := strings.Split(nodeNames, ",")

	if err := c.Storage.AddOrUpdatePodOfStorage(key, nodeNameSlice); err != nil {
		return err
	}

	keySlice := c.Storage.GetKeySliceOfStorage()
	klog.V(1).Infof("storage info: %v", keySlice)

	return nil
}
