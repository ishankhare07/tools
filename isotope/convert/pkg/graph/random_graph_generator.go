package graph

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/size"
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
	serviceGraph.Defaults = Defaults{
		ResponseSize: size.ByteSize(responseSize),
		RequestSize:  size.ByteSize(requestSize),
	}

	for i := 0; i < numberOfService; i++ {
		s := svc.Service{
			Name:           fmt.Sprintf("s%d", i),
			Type:           getRandomServiceType(generator),
			NumReplicas:    defaultNumReplicas,
			ClusterContext: getRandomCluster(listOfClusters, generator),
			Script:         getTargetRequestCommands(i, numberOfService),
		}

		serviceGraph.Services = append(serviceGraph.Services, s)
	}

	return *serviceGraph
}

func getLevel(nodeIndex int) int {
	height := math.Ceil(math.Log2(float64(nodeIndex+1)) - 1)
	return int(height)
}

func getAllNodesAtLevel(level, maxNodes int) []int {
	firstElement := math.Pow(float64(2), float64(level)) - 1
	numOfIterations := math.Pow(float64(2), float64(level))

	nodes := []int{}

	for i, j := firstElement, 0; j < int(numOfIterations) && int(i) < maxNodes; i, j = i+1, j+1 {
		nodes = append(nodes, int(i))
	}

	return nodes
}

func makeRequestCommand(child int) script.RequestCommand {
	requestCommand := script.RequestCommand{
		ServiceName: fmt.Sprintf("s%d", child),
	}

	return requestCommand
}

func getTargetRequestCommands(currentNode, numOfNodes int) script.Script {
	concurrentCommand := script.ConcurrentCommand{}

	maxHeight := getLevel(numOfNodes)
	currentLevel := getLevel(currentNode + 1)

	if currentLevel == maxHeight-1 {
		// connect to all nodes in the level below
		nodes := getAllNodesAtLevel(maxHeight, numOfNodes)
		for _, i := range nodes {
			concurrentCommand = append(concurrentCommand, makeRequestCommand(i))
		}
	} else {
		// connect to childs
		for i := 1; i < 3; i++ {
			child := 2*currentNode + 1
			if child <= numOfNodes {
				concurrentCommand = append(concurrentCommand, makeRequestCommand(child))
			}
		}
	}

	if len(concurrentCommand) == 0 {
		return script.Script{}
	}

	return script.Script{concurrentCommand}
}

func GetRandomFromRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func getRandomServiceType(generator RandomFromRange) svctype.ServiceType {
	//min := 1
	//max := 3 // max is not included in range
	//return svctype.ServiceType(generator(min, max))
	return svctype.ServiceType(svctype.ServiceHTTP)
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
