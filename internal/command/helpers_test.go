/*
 * Copyright © 2021 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package command

import (
	"fmt"

	camelkv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Shared test helpers

func createKamelet(kameletName string) *camelkv1alpha1.Kamelet {
	return createKameletInNamespace(kameletName, "default")
}

func createKameletInNamespace(kameletName string, namespace string) *camelkv1alpha1.Kamelet {
	return &camelkv1alpha1.Kamelet{
		TypeMeta: v1.TypeMeta{
			APIVersion: camelkv1alpha1.SchemeGroupVersion.String(),
			Kind:       camelkv1alpha1.KameletKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace:         namespace,
			Name:              kameletName,
			CreationTimestamp: v1.Now(),
			Labels: map[string]string{
				"camel.apache.org/kamelet.type":     "source",
				"camel.apache.org/kamelet.provider": "Community",
			},
			SelfLink: fmt.Sprintf("/apis/camel.apache.org/v1alpha1/namespaces/default/kamelets/%s", kameletName),
		},
		Spec: camelkv1alpha1.KameletSpec{
			Definition: &camelkv1alpha1.JSONSchemaProps{
				Title:       "Kamelet " + kameletName,
				Description: "Sample Kamelet source",
				Required:    []string{kameletName + "_prop"},
				Properties: map[string]camelkv1alpha1.JSONSchemaProps{
					kameletName + "_prop": {
						Type:        "string",
						Description: fmt.Sprintf("The %s required property", kameletName),
					},
					kameletName + "_optional": {
						Type:        "boolean",
						Description: fmt.Sprintf("The %s optional property", kameletName),
					},
				},
			},
		},
		Status: camelkv1alpha1.KameletStatus{
			Phase: camelkv1alpha1.KameletPhaseReady,
			Conditions: []camelkv1alpha1.KameletCondition{
				{
					Type:   camelkv1alpha1.KameletConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}
