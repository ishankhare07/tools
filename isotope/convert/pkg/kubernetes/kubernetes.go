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
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/tools/isotope/convert/pkg/graph/script"
	"k8s.io/apimachinery/pkg/util/intstr"
	"math/rand"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/tools/isotope/convert/pkg/consts"
	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
)

const (
	// ServiceGraphNamespace is the namespace all service graph related resources
	// (i.e. ConfigMap, Services, and Deployments) will reside in.
	ServiceGraphNamespace = "service-graph"

	DefaultGateway = "istio-system/ingressgateway"

	numConfigMaps          = 1
	numManifestsPerService = 2

	configVolume           = "config-volume"
	serviceGraphConfigName = "service-graph-config"
)

var (
	serviceGraphAppLabels       = map[string]string{"app": "service-graph"}
	serviceGraphNodeLabels      = map[string]string{"role": "service"}
	prometheusScrapeAnnotations = map[string]string{
		"prometheus.io/scrape": "true"}
)

// ManifestMap is a map of cluster context as keys to manifests
// as values which are basically array of strings each holding
// a single manifest corresponding to that cluster
type _manifestMap map[string][]string
type ManifestMap map[string]string

// ServiceGraphToKubernetesManifests converts a ServiceGraph to Kubernetes
// manifests.
func ServiceGraphToKubernetesManifests(
	serviceGraph graph.ServiceGraph,
	serviceNodeSelector map[string]string,
	serviceImage string,
	serviceMaxIdleConnectionsPerHost int,
	clientNodeSelector map[string]string,
	clientImage string,
	environmentName string) (ManifestMap, error) {

	manifestMap := make(_manifestMap)
	returnMap := make(ManifestMap)

	appendManifest := func(clusterName string, manifest interface{}) error {
		yamlDoc, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}

		if manifests, ok := manifestMap[clusterName]; !ok {
			// create the first entry
			manifestMap[clusterName] = []string{string(yamlDoc)}
		} else {
			manifestMap[clusterName] = append(manifests, string(yamlDoc))
		}

		return nil
	}

	addManifestToAllClusters := func(manifest interface{}) error {
		yamlDoc, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}

		for clusterContext, clusterManifests := range manifestMap {
			manifestMap[clusterContext] = append([]string{string(yamlDoc)}, clusterManifests...)
		}

		return nil
	}

	rand.Seed(time.Now().UTC().UnixNano())
	hasRbacPolicy := false
	for _, service := range serviceGraph.Services {
		k8sDeployment := makeDeployment(
			service, serviceNodeSelector, serviceImage,
			serviceMaxIdleConnectionsPerHost)
		innerErr := appendManifest(service.ClusterContext, k8sDeployment)
		if innerErr != nil {
			return nil, innerErr
		}

		k8sService := makeService(service)
		innerErr = appendManifest(service.ClusterContext, k8sService)
		if innerErr != nil {
			return nil, innerErr
		}

		serviceDependencies, dependencyErr := createServiceDependencies(service, serviceGraph)
		if dependencyErr != nil {
			return nil, dependencyErr
		}
		for _, serviceDependency := range serviceDependencies {
			innerErr = appendManifest(service.ClusterContext, serviceDependency)
			if innerErr != nil {
				return nil, innerErr
			}
		}

		// Only generates the RBAC rules when Istio is installed.
		if strings.EqualFold(environmentName, "ISTIO") {
			if service.NumRbacPolicies > 0 {
				hasRbacPolicy = true
				var i int32
				// Generates random RBAC rules for the service.
				for i = 0; i < service.NumRbacPolicies; i++ {
					addManifestToAllClusters(generateRbacPolicy(service, false /* allowAll */))
				}
				// Generates "allow-all" RBAC rule for the service.
				addManifestToAllClusters(generateRbacPolicy(service, true /* allowAll */))
			}

			if service.IsEntrypoint {
				virtualService := makeVirtualService(service.Name)
				innerErr = appendManifest(service.ClusterContext, virtualService)
				if innerErr != nil {
					return nil, innerErr
				}
			}
		}
	}

	fortioDeployment := makeFortioDeployment(
		clientNodeSelector,
		clientImage,
		serviceGraph.Global.IngressGatewayEndpoint,
		fmt.Sprintf("%s.%s.svc.cluster.local", serviceGraph.Services[0].Name, consts.ServiceGraphNamespace))
	if err := appendManifest(serviceGraph.Global.FortioCluster, fortioDeployment); err != nil {
		return nil, err
	}

	fortioService := makeFortioService()
	if err := appendManifest(serviceGraph.Global.FortioCluster, fortioService); err != nil {
		return nil, err
	}

	if hasRbacPolicy {
		addManifestToAllClusters(generateRbacConfig())
	}

	configMap, err := makeConfigMap(serviceGraph)
	if err != nil {
		return nil, err
	}

	if err := addManifestToAllClusters(configMap); err != nil {
		return nil, err
	}

	namespace := makeServiceGraphNamespace()
	if err := addManifestToAllClusters(namespace); err != nil {
		return nil, err
	}

	// yamlDocString := strings.Join(manifests, "---\n")
	for k, v := range manifestMap {
		returnMap[k] = strings.Join(v, "---\n")
	}

	return returnMap, nil
}

func createServiceDependencies(service svc.Service, serviceGraph graph.ServiceGraph) ([]apiv1.Service, error) {
	serviceDependencies := []apiv1.Service{}

	for _, commandScript := range service.Script {
		serviceDependency, err := generateDependencies(commandScript, serviceGraph, service.ClusterContext)
		if err != nil {
			return []apiv1.Service{}, err
		}
		serviceDependencies = append(serviceDependencies, serviceDependency...)
	}

	return serviceDependencies, nil
}

func generateDependencies(commandScript script.Command, serviceGraph graph.ServiceGraph, serviceCluster string) ([]apiv1.Service, error) {
	switch command := commandScript.(type) {
	case script.RequestCommand:
		downstreamService, err := serviceGraph.FindServiceByName(command.ServiceName)
		if err != nil {
			return nil, err
		}
		if serviceCluster != downstreamService.ClusterContext {
			k8sService := makeService(downstreamService)
			return []apiv1.Service{k8sService}, nil
		}
	case script.ConcurrentCommand:
		conccurentDependecies := []apiv1.Service{}
		for _, innerCommand := range command {
			dependencies, err := generateDependencies(innerCommand, serviceGraph, serviceCluster)
			if err != nil {
				return nil, err
			}
			conccurentDependecies = append(conccurentDependecies, dependencies...)
		}
		return conccurentDependecies, nil
	}
	return nil, nil
}

func combineLabels(a, b map[string]string) map[string]string {
	c := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}
	return c
}

func makeVirtualService(serviceName string) (v1alpha3.VirtualService) {
	serviceHost := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, consts.ServiceGraphNamespace)

	typeMeta := metav1.TypeMeta{
		Kind:       "VirtualService",
		APIVersion: "networking.istio.io/v1alpha3",
	}

	virtualServiceMeta := metav1.ObjectMeta{
		Name:      serviceName,
		Namespace: consts.ServiceGraphNamespace,
	}

	destination := &networkingv1alpha3.Destination{Host: serviceHost, Port: &networkingv1alpha3.PortSelector{Number: uint32(consts.ServicePort)}}

	spec := networkingv1alpha3.VirtualService{
		Gateways: []string{DefaultGateway},
		Hosts:    []string{serviceHost},
		Http:     []*networkingv1alpha3.HTTPRoute{
			{Route: []*networkingv1alpha3.HTTPRouteDestination{
				{
					Destination: destination,
				},
			}},
		},
	}

	return v1alpha3.VirtualService{
		TypeMeta:   typeMeta,
		ObjectMeta: virtualServiceMeta,
		Spec:       spec,
	}
}

func makeServiceGraphNamespace() (namespace apiv1.Namespace) {
	namespace.APIVersion = "v1"
	namespace.Kind = "Namespace"
	namespace.ObjectMeta.Name = consts.ServiceGraphNamespace
	namespace.ObjectMeta.Labels = map[string]string{"istio-injection": "enabled"}
	timestamp(&namespace.ObjectMeta)
	return
}

func makeConfigMap(
	graph graph.ServiceGraph) (configMap apiv1.ConfigMap, err error) {
	graphYAMLBytes, err := yaml.Marshal(graph)
	if err != nil {
		return
	}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.ObjectMeta.Name = serviceGraphConfigName
	configMap.ObjectMeta.Namespace = ServiceGraphNamespace
	configMap.ObjectMeta.Labels = serviceGraphAppLabels
	timestamp(&configMap.ObjectMeta)
	configMap.Data = map[string]string{
		consts.ServiceGraphConfigMapKey: string(graphYAMLBytes),
	}
	return
}

func makeService(service svc.Service) (k8sService apiv1.Service) {
	k8sService.APIVersion = "v1"
	k8sService.Kind = "Service"
	k8sService.ObjectMeta.Name = service.Name
	k8sService.ObjectMeta.Namespace = ServiceGraphNamespace
	k8sService.ObjectMeta.Labels = serviceGraphAppLabels
	timestamp(&k8sService.ObjectMeta)
	if service.Type.String() == "HTTP" {
		k8sService.Spec.Ports = []apiv1.ServicePort{
			{Port: consts.ServicePort, Name: consts.HttpServicePortName, TargetPort: intstr.FromInt(consts.ServicePort)}}
	} else {
		k8sService.Spec.Ports = []apiv1.ServicePort{
			{Port: consts.ServicePort, Name: consts.GrpcServicePortName, TargetPort: intstr.FromInt(consts.ServicePort)}}
	}
	k8sService.Spec.Selector = map[string]string{"name": service.Name}
	return
}

func makeDeployment(
	service svc.Service, nodeSelector map[string]string,
	serviceImage string, serviceMaxIdleConnectionsPerHost int) (
	k8sDeployment appsv1.Deployment) {
	k8sDeployment.APIVersion = "apps/v1"
	k8sDeployment.Kind = "Deployment"
	k8sDeployment.ObjectMeta.Name = service.Name
	k8sDeployment.ObjectMeta.Namespace = ServiceGraphNamespace
	k8sDeployment.ObjectMeta.Labels = serviceGraphAppLabels
	timestamp(&k8sDeployment.ObjectMeta)
	k8sDeployment.Spec = appsv1.DeploymentSpec{
		Replicas: &service.NumReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": service.Name,
			},
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: combineLabels(
					serviceGraphNodeLabels,
					map[string]string{
						"name": service.Name,
					}),
				Annotations: prometheusScrapeAnnotations,
			},
			Spec: apiv1.PodSpec{
				NodeSelector: nodeSelector,
				Containers: []apiv1.Container{
					{
						Name:  consts.ServiceContainerName,
						Image: serviceImage,
						Args: []string{
							fmt.Sprintf(
								"--max-idle-connections-per-host=%v",
								serviceMaxIdleConnectionsPerHost),
						},
						Env: []apiv1.EnvVar{
							{Name: consts.ServiceNameEnvKey, Value: service.Name},
						},
						VolumeMounts: []apiv1.VolumeMount{
							{
								Name:      configVolume,
								MountPath: consts.ConfigPath,
							},
						},
						Ports: []apiv1.ContainerPort{
							{
								ContainerPort: consts.ServicePort,
							},
						},
					},
				},
				Volumes: []apiv1.Volume{
					{
						Name: configVolume,
						VolumeSource: apiv1.VolumeSource{
							ConfigMap: &apiv1.ConfigMapVolumeSource{
								LocalObjectReference: apiv1.LocalObjectReference{
									Name: serviceGraphConfigName,
								},
								Items: []apiv1.KeyToPath{
									{
										Key:  consts.ServiceGraphConfigMapKey,
										Path: consts.ServiceGraphYAMLFileName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	timestamp(&k8sDeployment.Spec.Template.ObjectMeta)
	return
}

func timestamp(objectMeta *metav1.ObjectMeta) {
	objectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}
}
