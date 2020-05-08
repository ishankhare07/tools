// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/tools/isotope/convert/pkg/consts"
)

var fortioClientLabels = map[string]string{"app": "client"}

func makeFortioDeployment(
	nodeSelector map[string]string,
	clientImage string,
	ingressGatewayEndpoint string,
	downstreamServiceName string) (deployment appsv1.Deployment) {
	deployment.APIVersion = "apps/v1"
	deployment.Kind = "Deployment"
	deployment.ObjectMeta.Name = fmt.Sprintf("client-%s", downstreamServiceName)
	deployment.ObjectMeta.Labels = fortioClientLabels
	timestamp(&deployment.ObjectMeta)
	deployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": fmt.Sprintf("client-%s", downstreamServiceName)},
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": fmt.Sprintf("client-%s", downstreamServiceName)},
				Annotations: map[string]string{
					"sidecar.istio.io/inject": "false",
				},
			},
			Spec: apiv1.PodSpec{
				NodeSelector: nodeSelector,
				Containers: []apiv1.Container{
					{
						Name:  "fortio-client",
						Image: clientImage,
						Args: []string{
							"load",
							"-c",
							"8",
							"-qps",
							"200",
							"-t",
							"0",
							"-r",
							"0.0001",
							"-H",
							fmt.Sprintf("Host: %s.%s.svc.cluster.local", downstreamServiceName, consts.ServiceGraphNamespace),
							fmt.Sprintf("http://%s", ingressGatewayEndpoint),
						},
						Ports: []apiv1.ContainerPort{
							{
								ContainerPort: consts.ServicePort,
							},
							{
								ContainerPort: consts.FortioMetricsPort,
							},
						},
					},
				},
			},
		},
	}
	timestamp(&deployment.Spec.Template.ObjectMeta)
	return
}

func makeFortioService(downstreamServiceName string) (service apiv1.Service) {
	service.APIVersion = "v1"
	service.Kind = "Service"
	service.ObjectMeta.Name = fmt.Sprintf("client-%s", downstreamServiceName)
	service.ObjectMeta.Labels = fortioClientLabels
	service.ObjectMeta.Annotations = prometheusScrapeAnnotations
	timestamp(&service.ObjectMeta)
	service.Spec.Ports = []apiv1.ServicePort{{Port: consts.ServicePort}}
	service.Spec.Selector = map[string]string{"app": fmt.Sprintf("client-%s", downstreamServiceName)}
	return
}
