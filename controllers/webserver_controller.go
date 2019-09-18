/*
Copyright 2019 cnc-demo.

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
	"fmt"
	"reflect"
	run "runtime"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serversv1beta1 "cnc-demo/api/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// WebServerReconciler reconciles a WebServer object
type WebServerReconciler struct {
	client.Client
	Log    logr.Logger
	scheme *runtime.Scheme
	mutex  sync.Mutex
}

// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webservers/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=service,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=service/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch

func (r *WebServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("webserver", req.NamespacedName)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	stackb := make([]byte, 1<<20)
	run.Stack(stackb, true)
	stack := strings.Split(string(stackb), "\n")[0]
	r.Log.V(1).Info("WEB SERVER RECONCILE START: " + stack)
	defer r.Log.V(1).Info("WEB SERVER RECONCILE END  : " + stack)

	instance := &serversv1beta1.WebServer{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := define_deployment(instance)
	r.Log.V(1).Info("VAR", "INSTANCE", instance)

	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		r.Log.V(1).Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		return reconcile.Result{}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	r.Log.V(1).Info("VAR", "found", found)

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		r.Log.V(1).Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Reconcile of SERVICE
	service := define_service(instance)
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	foundSvc := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		r.Log.V(1).Info("Creating Service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		return reconcile.Result{}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO not exhaustive check
	if !reflect.DeepEqual(service.Spec.Ports[0].Port, foundSvc.Spec.Ports[0].Port) {
		foundSvc.Spec.Ports = service.Spec.Ports
		r.Log.V(1).Info("Updating Service", "namespace", service.Namespace, "name", service.Name)
		err = r.Update(context.TODO(), foundSvc)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *WebServerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	r.scheme = mgr.GetScheme()

	// Create a new controller
	c, err := controller.New("webserver-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to WebServer
	err = c.Watch(&source.Kind{Type: &serversv1beta1.WebServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by WebServer - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &serversv1beta1.WebServer{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &serversv1beta1.WebServer{},
	})

	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&serversv1beta1.WebServer{}).
		Complete(r)
}

func define_deployment(instance *serversv1beta1.WebServer) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
							Command: []string{
								"/bin/sh",
								"-c",
								fmt.Sprintf("echo %s > /usr/share/nginx/html/index.html; nginx -g 'daemon off;'", instance.Spec.Content),
							},
						},
					},
				},
			},
		},
	}
}

func define_service(instance *serversv1beta1.WebServer) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-service",
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "http-port",
					Port:       instance.Spec.Port.HTTP,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"deployment": instance.Name + "-deployment",
			},
			// Type: corev1.ServiceTypeLoadBalancer,
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}
