package graph

import (
	"fmt"
	"math/rand"
	"time"

	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

const defaultNumReplicas = 6

type RandomFromRange func(min, max int) int

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateRandomServiceGraph(numberOfService int,
	requestSize int,
	responseSize int,
	listOfClusters []string,
	ingressGatewayEndpoint string,
	generator RandomFromRange) ServiceGraph {
	serviceGraph := new(ServiceGraph)
	serviceGraph.Global = generateServiceDefaults(listOfClusters, ingressGatewayEndpoint, generator)

	for i := 0; i < numberOfService; i++ {
		s := svc.Service{
			Name:           fmt.Sprintf("s%d", i),
			Type:           getRandomServiceType(generator),
			NumReplicas:    defaultNumReplicas,
			ClusterContext: getRandomCluster(listOfClusters, generator),
			Script: script.Script{
				getTargetRequestCommands(i, numberOfService),
			},
		}

		serviceGraph.Services = append(serviceGraph.Services, s)
	}

	return *serviceGraph
}

func getTargetRequestCommands(serviceToSkip, numOfServices int) []script.RequestCommand {
	requestCommands := []script.RequestCommand{}

	for i := 0; i < numOfServices; i++ {
		if i != serviceToSkip {
			requestCommand := script.RequestCommand{
				ServiceName: fmt.Sprintf("s%d", i),
			}

			requestCommands = append(requestCommands, requestCommand)
		}
	}

	return requestCommands
}

func GetRandomFromRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func getRandomServiceType(generator RandomFromRange) svctype.ServiceType {
	min := 1
	max := 3 // max is not included in range
	return svctype.ServiceType(generator(min, max))
	// return svctype.ServiceType(2)
}

func generateServiceDefaults(listOfClusters []string, ingressGatewayEndpoint string, generator RandomFromRange) ServiceDefaults {
	serviceDefaults := new(ServiceDefaults)
	serviceDefaults.FortioCluster = getRandomCluster(listOfClusters, generator)
	serviceDefaults.IngressGatewayEndpoint = ingressGatewayEndpoint

	return *serviceDefaults
}

func getRandomCluster(listOfClusters []string, generator RandomFromRange) string {
	min := 0
	max := len(listOfClusters)
	return listOfClusters[generator(min, max)]
}
