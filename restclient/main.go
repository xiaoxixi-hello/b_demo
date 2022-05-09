package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	// 1. 获取kubeconfig 连接集群
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			klog.Error(err, "get kubeconfig failed")
		}
		config = inClusterConfig
	}

	// 2. 实例化restClient  /api/v1/namespaces/{namespace}/pods
	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	// 序列化工具
	config.NegotiatedSerializer = scheme.Codecs

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		klog.Error(err)
	}

	// 3. 获取pod列表
	result := &corev1.PodList{}
	if err := restClient.Get().
		// 指定ns
		Namespace("kube-system").
		// 指定resource
		Resource("pods").
		// 指定大小和序列化工具
		VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec).
		// 写入提前准备好的list中
		Do(context.TODO()).Into(result); err != nil {
		klog.Error(err)
	}

	// 4. 格式化输出
	fmt.Printf("Namespace\t Status\t\t Name\n")
	// 每个pod都打印Namespace、Status.Phase、Name三个字段
	for _, d := range result.Items {
		fmt.Printf("%v\t %v\t %v\n",
			d.Namespace,
			d.Status.Phase,
			d.Name)
	}
}
