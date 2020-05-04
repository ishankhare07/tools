package graph

import (
	"math/rand"
	"time"
)

func GenerateRandomServiceGraph(numberOfService int, requestSize int, responseSize int, listOfClusters []string, ingressGatewayEndpoint string) ServiceGraph {
	serviceGraph := new(ServiceGraph)
	serviceGraph.Global = generateServiceDefaults(listOfClusters, ingressGatewayEndpoint)

	return *serviceGraph
}

func generateServiceDefaults(listOfClusters []string, ingressGatewayEndpoint string) ServiceDefaults {
	serviceDefaults := new(ServiceDefaults)
	serviceDefaults.FortioCluster = getRandomCluster(listOfClusters)
	serviceDefaults.IngressGatewayEndpoint = ingressGatewayEndpoint

	return *serviceDefaults
}

func getRandomCluster(listOfClusters []string) string {
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := len(listOfClusters)
	return listOfClusters[rand.Intn(max - min) + min]
}
