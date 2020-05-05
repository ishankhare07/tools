package graph

import (
	"reflect"
	"testing"

	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

func MockGenerator(min, max int) int {
	return 1
}

func TestGenerateRandomServiceGraph(t *testing.T) {
	numberOfService := 2
	requestSize := 5
	responseSize := 5
	clusterList := []string{"cluster0", "cluster1"}
	ingressEndpoint := "x.x.x.x"

	svcGraph := GenerateRandomServiceGraph(numberOfService, requestSize, responseSize, clusterList, ingressEndpoint, MockGenerator)

	expectedGraph := ServiceGraph{
		Services: []svc.Service{
			{
				Name:           "s0",
				Type:           svctype.ServiceHTTP,
				NumReplicas:    6,
				ClusterContext: "cluster1",
				Script: script.Script{
					[]script.RequestCommand{
						{
							ServiceName: "s1",
						},
					},
				},
			},
			{
				Name:           "s1",
				Type:           svctype.ServiceHTTP,
				NumReplicas:    6,
				ClusterContext: "cluster1",
				Script: script.Script{
					[]script.RequestCommand{
						{
							ServiceName: "s0",
						},
					},
				},
			},
		},
		Global: ServiceDefaults{
			FortioCluster:          "cluster1",
			IngressGatewayEndpoint: ingressEndpoint,
		},
	}

	if len(svcGraph.Services) != numberOfService {
		t.Errorf("Not correct number of services. expected %d, got %d", numberOfService, len(svcGraph.Services))
	}

	if !reflect.DeepEqual(svcGraph.Services, expectedGraph.Services) {
		t.Errorf("Services do not match")
	}
}