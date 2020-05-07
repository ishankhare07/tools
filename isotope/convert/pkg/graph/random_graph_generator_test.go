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
	numberOfService := 5
	requestSize := 5
	responseSize := 5
	clusterList := []string{"cluster0", "cluster1"}
	ingressEndpoint := "x.x.x.x"

	svcGraph := GenerateRandomServiceGraph(numberOfService, 5, requestSize, responseSize, clusterList, ingressEndpoint, MockGenerator)

	expectedGraph := ServiceGraph{
		Services: []svc.Service{
			{
				Name:           "s0",
				Type:           svctype.ServiceHTTP,
				NumReplicas:    6,
				ClusterContext: "cluster1",
				Script: script.Script{
					script.ConcurrentCommand{
						script.RequestCommand{
							ServiceName: "s1",
						},
						script.RequestCommand{
							ServiceName: "s2",
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
					script.ConcurrentCommand{
						script.RequestCommand{
							ServiceName: "s3",
						},
						script.RequestCommand{
							ServiceName: "s4",
						},
					},
				},
			},
			{
				Name:           "s2",
				Type:           svctype.ServiceHTTP,
				NumReplicas:    6,
				ClusterContext: "cluster1",
				Script: script.Script{
					script.ConcurrentCommand{
						script.RequestCommand{
							ServiceName: "s3",
						},
						script.RequestCommand{
							ServiceName: "s4",
						},
					},
				},
			},
			{
				Name:           "s3",
				Type:           svctype.ServiceHTTP,
				NumReplicas:    6,
				ClusterContext: "cluster1",
				Script:         script.Script{},
			},
			{
				Name:           "s4",
				Type:           svctype.ServiceHTTP,
				NumReplicas:    6,
				ClusterContext: "cluster1",
				Script:         script.Script{},
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
		t.Errorf("Services do not match\n%#v\n%#v", expectedGraph.Services, svcGraph.Services)
	}
}

func TestGetAllNodesAtLevel(t *testing.T) {
	nodes := getAllNodesAtLevel(15, 3)
	expected := []int{}

	if !reflect.DeepEqual(nodes, expected) {
		t.Errorf("Not correct list of nodes, expected %#v, got %#v", expected, nodes)
	}
}
