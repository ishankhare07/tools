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

package graph

import (
	"fmt"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
)

// ServiceGraph describes a set of services which mock a service-oriented
// architecture.
type ServiceGraph struct {
	Global   ServiceDefaults `json:"global,omitempty"`
	Services []svc.Service   `json:"services"`
	Defaults Defaults        `json:"defaults"`
}

type ServiceDefaults struct {
	IngressGatewayEndpoint string   `json:"ingress_gateway_endpoint,omitempty"`
	ControlPlaneClusters   []string `json:"control_plane_clusters,omitempty"`
}

func (serviceGraph ServiceGraph) FindServiceByName(serviceName string) (svc.Service, error) {
	for _, service := range serviceGraph.Services {
		if service.Name == serviceName {
			return service, nil
		}
	}

	return svc.Service{}, fmt.Errorf("%s is not found", serviceName)
}
