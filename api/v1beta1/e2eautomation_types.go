/*
Copyright 2023.

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
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Endpoint struct {
	Id   string `json:"id"`
	Url  string `json:"url"`
	User User   `json:"user"`
}

type Settings struct {
	Endpoint                Endpoint `json:"endpoint"`
	RetryPolicy             int      `json:"retryPolicy,omitempty"`
	OpcuaServerDiscoveryUrl string   `json:"opcuaServerDiscoveryUrl"`
}

// E2EAutomationSpec defines the desired state of E2EAutomation
type E2EAutomationSpec struct {
	Id       string   `json:"id"`
	ImageId  string   `json:"imageId"`
	Schedule string   `json:"schedule"`
	Settings Settings `json:"settings"`
}

// E2EAutomationStatus defines the observed state of E2EAutomation
type E2EAutomationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// E2EAutomation is the Schema for the e2eautomations API
type E2EAutomation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   E2EAutomationSpec   `json:"spec,omitempty"`
	Status E2EAutomationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// E2EAutomationList contains a list of E2EAutomation
type E2EAutomationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []E2EAutomation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&E2EAutomation{}, &E2EAutomationList{})
}
