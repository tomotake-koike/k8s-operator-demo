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
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"net/http"
	"reflect"
	run "runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serversv1beta1 "cnc-demo/api/v1beta1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// WebManagerReconciler reconciles a WebManager object
type WebManagerReconciler struct {
	client.Client
	Log            logr.Logger
	scheme         *runtime.Scheme
	instance       serversv1beta1.WebManager
	mutex          sync.Mutex
	NamespacedName types.NamespacedName
}

// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webmanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webmanagers/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=service,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=service/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch

func (r *WebManagerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("webmanager", req.NamespacedName)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.NamespacedName = req.NamespacedName

	stackb := make([]byte, 1<<20)
	run.Stack(stackb, true)
	stack := strings.Split(string(stackb), "\n")[0]
	r.Log.V(1).Info("RECONCILE START: " + stack)
	defer r.Log.V(1).Info("RECONCILE END  : " + stack)

	instance := &serversv1beta1.WebManager{}
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

	instanceName := req.NamespacedName.String()

	// TODO: invalid Re-reconcile same instance
	if reflect.DeepEqual(r.instance, *instance) {
		r.Log.V(1).Info("[END] SAME INSTANCE", "Before:", r.instance, "After:", *instance)
		return reconcile.Result{}, err
	}

	// Auto Delete
	if instance.Spec.WebSets <= instance.Status.CreatedSets {

		if instance.Spec.AutoDelete {
			r.Log.V(1).Info("AUTO DELETE!!")
			time.Sleep(instance.Spec.WaitForFarewell * time.Millisecond)
			err = r.Delete(context.TODO(), instance)
			if err != nil {
				r.Log.V(1).Info("ERR "+stack, "err", err, "Failed to delete :", instance)
			}
		}
		return reconcile.Result{}, err
	}

	// Get List of WebServers
	webServerList := &serversv1beta1.WebServerList{}
	ctx := context.TODO()
	err = r.List(ctx, webServerList)
	if err != nil {
		r.Log.V(1).Info("ERR", "err", err, "Failed to list for: ", instanceName)
		return reconcile.Result{}, err
	}

	var dependents []serversv1beta1.WebServer
	for _, item := range webServerList.Items {
		for _, owner := range item.GetOwnerReferences() {
			if item.Name == instance.Name {
				dependents = append(dependents, item)
			} else {
				r.Log.V(1).Info("DEPEND UNMATCH "+stack, "OWNER", owner.Name, "INSTANCE", item.Name)
			}
		}
	}

	webServers_list := len(dependents)

	r.Log.V(1).Info("VAR", "INSTANCE", instance, "WebServers_list: ", webServers_list)

	if instance.Status.State == 0 {
		// Create WebServer
		time.Sleep(instance.Spec.CreateIntervalMillis * time.Millisecond)

		webserver := define_webserver(instance)
		if err := controllerutil.SetControllerReference(instance, webserver, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.V(1).Info("CREATE WebServer : " + webserver.Name)
		if err := r.Create(context.TODO(), webserver); err != nil {
			r.Log.V(1).Info("ERR "+stack, "err", err, "Failed to create WebServer:", webserver.Name)
			return reconcile.Result{}, err
		}
		instance.Status.State = 1
		instance.Status.LastUpdate = time.Now().String()
		r.Log.V(1).Info("UPDATE WebManager 0 -> 1 : " + instance.Name)
		if err := r.Update(context.TODO(), instance); err != nil {
			r.Log.V(1).Info("ERR  "+stack, "err", err, "Failed to update from 0 to 1 :", instanceName)
			return reconcile.Result{}, err
		}

	} else if instance.Status.State == 1 {
		// Update WebServer Status : CreatedSets
		instance.Status.CreatedSets++
		instance.Status.State = 2
		instance.Status.LastUpdate = time.Now().String()
		r.Log.V(1).Info("UPDATE WebManager 1 -> 2 : " + instance.Name)
		if err := r.Update(context.TODO(), instance); err != nil {
			r.Log.V(1).Info("ERR  "+stack, "err", err, "Failed to update from 1 to 2 :", instanceName)
			return reconcile.Result{}, err
		}

	} else if instance.Status.State == 2 {
		// Update WebServer Status : State 2 -> 0
		instance.Status.State = 0
		instance.Status.LastUpdate = time.Now().String()
		r.Log.V(1).Info("UPDATE WebManager 2 -> 0 : " + instance.Name)
		if err := r.Update(context.TODO(), instance); err != nil {
			r.Log.V(1).Info("ERR  "+stack, "err", err, "Failed to update from 2 to 0 :", instanceName)
			return reconcile.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

func (r *WebManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	r.scheme = mgr.GetScheme()

	// Create a new controller
	c, err := controller.New("webmanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to WebManager
	err = c.Watch(&source.Kind{Type: &serversv1beta1.WebManager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a WebServer created by WebManager - change this for objects you create
	err = c.Watch(&source.Kind{Type: &serversv1beta1.WebServer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &serversv1beta1.WebManager{},
	})
	if err != nil {
		return err
	}

	go start_api_server(r)

	return ctrl.NewControllerManagedBy(mgr).
		For(&serversv1beta1.WebManager{}).
		Complete(r)
}

func define_webserver(instance *serversv1beta1.WebManager) *serversv1beta1.WebServer {
	return &serversv1beta1.WebServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-webserver-" + fmt.Sprintf("%04d", instance.Status.CreatedSets),
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.WebServer,
	}
}

func start_api_server(r *WebManagerReconciler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		r.mutex.Lock()
		defer r.mutex.Unlock()

		if req.Method == "GET" || req.Method == "REFILL" {

			instance := &serversv1beta1.WebManager{}
			err := r.Get(context.TODO(), r.NamespacedName, instance)
			if err != nil {
				if errors.IsNotFound(err) {
					fmt.Fprintf(w, "Object not found, return.  Created objects are automatically garbage collected.\n")
					fmt.Fprintf(w, "For additional cleanup logic use finalizers.\n")
					return
				}
				fmt.Fprintf(w, "Error reading the object - requeue the request.")
				return
			}

			if req.Method == "REFILL" {

				// REFILL
				instance.Spec.WebSets++
				// Update WebServer Status : State 2 -> 0
				instance.Status.State = 0
				instance.Status.LastUpdate = time.Now().String()
				r.Log.V(1).Info("REFILL UPDATE WebManager 2 -> 0 : " + instance.Name)
				if err := r.Update(context.TODO(), instance); err != nil {
					fmt.Fprintf(w, "Refill Error : Failed to update from 2 to 0 : %s", instance.Name)
					r.Log.V(1).Info("ERR  ", err, "Failed to update from 2 to 0 :", instance.Name)
					return
				}
			}

			prettyJSON, jsonerr := json.MarshalIndent(instance, "", "    ")
			if jsonerr != nil {
				fmt.Fprintf(w, "%+v\n", instance)
			} else {
				fmt.Fprintf(w, "%s\n", prettyJSON)
			}

		} else {

			fmt.Fprintf(w, "NamespacedName: %+v", r.NamespacedName)
			fmt.Fprintf(w, "Received Unsuppoted Method : %s \n", req.Method)

		}

	})
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":10080")
}
