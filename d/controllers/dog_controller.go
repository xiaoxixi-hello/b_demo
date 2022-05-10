/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	dv1 "github.com/ylinyang/b_demo/d/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DogReconciler reconciles a Dog object
type DogReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=d.y.demo.io,resources=dogs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=d.y.demo.io,resources=dogs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=d.y.demo.io,resources=dogs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *DogReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. 只有cr存在的时候 才会controller才能走的下去
	instance := &dv1.Dog{}
	err := r.Get(ctx, req.NamespacedName, instance)

	// 2. 判断对应cr的deploy是否存在
	deploy := &appsv1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, deploy)
	if err == nil {
		log.Info("deploy is exists")
		return ctrl.Result{}, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "get deploy is failed")
		return ctrl.Result{}, err
	}
	log.Info("deploy is not found")
	// 3. 实例化一个deploy的数据结构
	deploy = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "dog-x",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "dog-x",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "dog-x",
						Image: instance.Spec.Images,
					}},
				},
			},
		},
	}

	// 4. 设置deploy 与 cr的关联, 需要先定义deploy在关联对象，且命名空间要一致
	log.Info("set reference")
	if err := ctrl.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		log.Error(err, "set reference about dogs and deploy is failed")
		return ctrl.Result{}, err
	}

	// 5. 开始创建deploy
	log.Info("start create deploy")
	if err := r.Client.Create(ctx, deploy); err != nil {
		log.Error(err, "create deploy is failed")
		return ctrl.Result{}, err
	}
	log.Info("create is ok")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dv1.Dog{}).Complete(r)
}
