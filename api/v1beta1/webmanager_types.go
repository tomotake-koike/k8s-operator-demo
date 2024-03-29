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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebManagerSpec defines the desired state of WebManager
type WebManagerSpec struct {
	WebSets              int           `json:"webSets"`
	AutoDelete           bool          `json:"autoDelete,omitempty"`
	CreateIntervalMillis time.Duration `json:"intervalMillis"`
	WaitForFarewell      time.Duration `json:"waitForFarewellMillis"`

	WebServer WebServerSpec `json:"webServer"`
}

// WebManagerStatus defines the observed state of WebManager
type WebManagerStatus struct {
	CreatedSets int    `json:"createdSets"`
	Achievement bool   `json:"Achievement"`
	LastUpdate  string `json:"lastUpdate"`
	State       int    `json:"state"`
}

// +kubebuilder:object:root=true

// WebManager is the Schema for the webmanagers API
type WebManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebManagerSpec   `json:"spec,omitempty"`
	Status WebManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebManagerList contains a list of WebManager
type WebManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebManager{}, &WebManagerList{})
}
