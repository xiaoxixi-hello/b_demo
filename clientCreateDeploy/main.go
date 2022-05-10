package main

import (
    "context"
    "log"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

func main() {
    // 1. 获取kubeconfig文件
    config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        inClusterConfig, err := rest.InClusterConfig()
        if err != nil {
            log.Panicln("get kubeconfig failed -- ", err)
        }
        config = inClusterConfig
    }

    // 2. 创建clientset客户端
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Panicln(err)
    }

    // 3. 创建deploy
    ns := "default"
    labels := map[string]string{"app": "nginx"}
    deploy := &appsv1.Deployment{
        TypeMeta: metav1.TypeMeta{},
        ObjectMeta: metav1.ObjectMeta{
            Namespace: ns,
            Name:      "nginx",
        },
        Spec: appsv1.DeploymentSpec{
            Selector: &metav1.LabelSelector{
                MatchLabels: labels,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: labels,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Name:  "nginx-demo",
                        Image: "nginx",
                    }},
                },
            },
        },
        Status: appsv1.DeploymentStatus{},
    }
    _, err = clientset.AppsV1().Deployments(ns).Create(context.TODO(), deploy, metav1.CreateOptions{})
    if err != nil {
        log.Panicln("create deploy failed -- ", err)
    }
}
