package main

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	// 1. 获取配置文件 默认目录 ~/.kube/config
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		// 使用集群内部的配置文件 /var/run/secrets/kubernetes.io/serviceaccount/
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			klog.Error(err)
		}
		config = inClusterConfig
	}
	// 打印一下config文件里面的host
	fmt.Println(config.Host)
}

/*
https://kubernetes.docker.internal:6443
*/
