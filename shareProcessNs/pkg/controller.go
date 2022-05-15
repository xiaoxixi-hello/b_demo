package pkg

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informersv1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	clientset kubernetes.Interface

	deploymentLister listersv1.DeploymentLister
	deploymentSynced cache.InformerSynced
	workqueue        workqueue.RateLimitingInterface
}

func NewController(clientset kubernetes.Interface, deploymentInformer informersv1.DeploymentInformer) *Controller {
	c := &Controller{
		clientset:        clientset,
		deploymentLister: deploymentInformer.Lister(),
		deploymentSynced: deploymentInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "shareProcessNs"),
	}

	log.Println("Setting up event handlers")
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("--------------------------------------add func")
			c.queue(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Println("--------------------------------------add update")
			c.queue(newObj)
		},
		DeleteFunc: nil,
	})
	return c
}

func (c *Controller) Run(stopCh chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	log.Println("Starting shareProcessNs controller")

	log.Println("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Println("Starting workers")

	go wait.Until(c.runWorker, time.Second, stopCh)

	log.Println("Started workers")
	<-stopCh
	log.Println("shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {

	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// obj active
	err := func(obj interface{}) error {
		// 处理该obj后需要done
		defer c.workqueue.Done(obj)

		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// 不希望该obj重新入队
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing %s: %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		log.Printf("Successfully synced %s \n", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) queue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("getting key from cache %s\n", err.Error()))
	}
	c.workqueue.Add(key)
}

func (c *Controller) syncHandler(key string) error {
	log.Println(key + " 1. get ns and name")
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	log.Println(key + " 2. get shell and annotations")
	deployment, err := c.deploymentLister.Deployments(namespace).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if _, ok := deployment.GetAnnotations()["shell"]; !ok {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == "shell" {
				log.Println(key, " 2.1 find shell, but annotations is nil, start delete shell")
				deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers[:i],
					deployment.Spec.Template.Spec.Containers[i+1:]...)
				deployment.Spec.Template.Spec.ShareProcessNamespace = boolPtr(false)
				if _, err := c.clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == "shell" {
			log.Println(key, " 2.1 find shell, not create")
			return nil
		}
	}
	log.Println(key + " 3. add container")
	container := v1.Container{
		Name:  "shell",
		Image: "busybox:1.28",
		SecurityContext: &v1.SecurityContext{
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{"SYS_PTRACE"},
			},
		},
		Stdin: true,
		TTY:   true,
	}
	deployment.Spec.Template.Spec.ShareProcessNamespace = boolPtr(true)
	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container)
	if _, err := c.clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
		return err
	}
	log.Println(key + "\t4. fix ok")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }
