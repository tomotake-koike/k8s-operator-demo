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

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serversv1beta1 "cnc-demo/api/v1beta1"
)

// WebManagerReconciler reconciles a WebManager object
type WebManagerReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webmanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servers.cnc.demo,resources=webmanagers/status,verbs=get;update;patch

func (r *WebManagerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("webmanager", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *WebManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serversv1beta1.WebManager{}).
		Complete(r)
}
