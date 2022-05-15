package main

import (
	"log"
	"time"

	"gihtub.com/ylinyang/b_demo/shareProcessNs/pkg"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 1. get kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Panicln("get kubeconfig is failed, ", err)
		}
		config = inClusterConfig
	}

	// 2. clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicln(err)
	}

	// 3. informer
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	deploymentInformer := sharedInformerFactory.Apps().V1().Deployments()

	ch := make(chan struct{})
	controller := pkg.NewController(clientset, deploymentInformer)

	sharedInformerFactory.Start(ch)
	if err := controller.Run(ch); err != nil {
		log.Panicln(err)
	}
}
