/*
 * Copyright 2018. IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package learner

import (
	v1beta1 "k8s.io/api/apps/v1beta1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreatePodSpec ...
func CreatePodSpec(containers []v1core.Container, volumes []v1core.Volume, labels map[string]string, nodeSelector map[string]string, imagePullSecret []v1core.LocalObjectReference, nodeAffinity *v1core.NodeAffinity, gpuToleration []v1core.Toleration) v1core.PodTemplateSpec {
	labels["service"] = "dlaas-learner" //label that denies ingress/egress
	automountSeviceToken := false
	return v1core.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
			Annotations: map[string]string{
				"scheduler.alpha.kubernetes.io/tolerations": `[ { "key": "dedicated", "operator": "Equal", "value": "gpu-task" } ]`,
				"scheduler.alpha.kubernetes.io/nvidiaGPU":   `{ "AllocationPriority": "Dense" }`,
			},
		},
		Spec: v1core.PodSpec{
			Containers:                   containers,
			Volumes:                      volumes,
			ImagePullSecrets:             imagePullSecret,
			Tolerations:                  gpuToleration,
			NodeSelector:                 nodeSelector,
			AutomountServiceAccountToken: &automountSeviceToken,
			Affinity: &v1core.Affinity{
				NodeAffinity: nodeAffinity,
			},
		},
	}
}

//CreateStatefulSetSpecForLearner ...
func CreateStatefulSetSpecForLearner(name, servicename string, replicas int, podTemplateSpec v1core.PodTemplateSpec) *v1beta1.StatefulSet {
	var replicaCount = int32(replicas)
	revisionHistoryLimit := int32(0) //https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#clean-up-policy

	return &v1beta1.StatefulSet{

		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: podTemplateSpec.Labels,
		},
		Spec: v1beta1.StatefulSetSpec{
			ServiceName:          servicename,
			Replicas:             &replicaCount,
			Template:             podTemplateSpec,
			RevisionHistoryLimit: &revisionHistoryLimit, //we never rollback these
			//PodManagementPolicy: v1beta1.ParallelPodManagement, //using parallel pod management in stateful sets to ignore the order. not sure if this will affect the helper pod since any pod in learner can come up now
		},
	}
}

//CreateServiceSpec ... this service will govern the statefulset
func CreateServiceSpec(name string, trainingID string) *v1core.Service {

	return &v1core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"training_id": trainingID,
			},
		},
		Spec: v1core.ServiceSpec{
			Selector: map[string]string{"training_id": trainingID},
			Ports: []v1core.ServicePort{
				v1core.ServicePort{
					Name:     "ssh",
					Protocol: v1core.ProtocolTCP,
					Port:     22,
				},
				v1core.ServicePort{
					Name:     "tf-distributed",
					Protocol: v1core.ProtocolTCP,
					Port:     2222,
				},
			},
			ClusterIP: "None",
		},
	}
}
