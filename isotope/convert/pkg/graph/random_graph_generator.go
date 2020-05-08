package graph

import (
	"fmt"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
	"math"
	"math/rand"
	"time"

	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/size"
)

const defaultNumReplicas = 6

type RandomFromRange func(min, max int) int

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateRandomServiceGraph(numberOfServices int,
	subTreeHeight int,
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

	for subTree := 0; subTree < int(math.Ceil(float64(numberOfServices)/float64(getMaxNodesOfSubtree(subTreeHeight)))); subTree++ {
		for node := 0; node < getMaxNodesOfSubtree(subTreeHeight); node++ {
			isRootNode := node == subTree * getMaxNodesOfSubtree(subTreeHeight)
			s := svc.Service{
				Name:            fmt.Sprintf("s%d", subTree*getMaxNodesOfSubtree(subTreeHeight)+node),
				Type:            svctype.ServiceType(svctype.ServiceHTTP),
				NumReplicas:     defaultNumReplicas,
				IsEntrypoint:    isRootNode,
				Script:          getTargetRequestCommands(node, getMaxNodesOfSubtree(subTreeHeight), subTree*getMaxNodesOfSubtree(subTreeHeight)),
				ClusterContext:  getRandomCluster(listOfClusters, generator),
			}

			serviceGraph.Services = append(serviceGraph.Services, s)
		}
	}

	return *serviceGraph
}

func getMaxNodesOfSubtree(height int) int {
	return int(math.Pow(float64(2), float64(height+1)) - 1)
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

func getTargetRequestCommands(currentNode,
	numOfNodes,
	startLabel int) script.Script {
	concurrentCommand := script.ConcurrentCommand{}

	maxHeight := getLevel(numOfNodes)
	currentLevel := getLevel(currentNode + 1)

	if currentLevel == maxHeight-1 {
		// connect to all nodes in the level below
		nodes := getAllNodesAtLevel(maxHeight, numOfNodes)
		for _, i := range nodes {
			label := startLabel + i
			concurrentCommand = append(concurrentCommand, makeRequestCommand(label))
		}
	} else {
		// connect to childs
		for i := 1; i < 3; i++ {
			child := 2*currentNode + i
			if child < numOfNodes {
				label := startLabel + child
				concurrentCommand = append(concurrentCommand, makeRequestCommand(label))
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
