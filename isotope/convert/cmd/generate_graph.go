/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"strings"

	"istio.io/tools/isotope/convert/pkg/output"

	"github.com/ghodss/yaml"

	"github.com/spf13/cobra"
	"istio.io/tools/isotope/convert/pkg/graph"
)

// generateGraphCmd represents the generateGraph command
var generateGraphCmd = &cobra.Command{
	Use:   "generate-graph [output.yaml]",
	Short: "Generate isotope config yaml",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		numberOfServices, err := cmd.PersistentFlags().GetInt("number-of-services")
		exitIfError(err)
		maxSubtreeSize, err := cmd.PersistentFlags().GetInt("subtree-size")
		exitIfError(err)
		requestSize, err := cmd.PersistentFlags().GetInt("request-size")
		exitIfError(err)
		responseSize, err := cmd.PersistentFlags().GetInt("response-size")
		exitIfError(err)
		clusters, err := cmd.PersistentFlags().GetString("cluster-list")
		exitIfError(err)
		clusterList := strings.Split(clusters, ",")
		ingressGatewayEndpoint, err := cmd.PersistentFlags().GetString("ingress-gateway-endpoint")
		exitIfError(err)

		targetFilename := args[0]

		svcGraph := graph.GenerateRandomServiceGraph(numberOfServices, maxSubtreeSize, requestSize, responseSize, clusterList, ingressGatewayEndpoint, graph.GetRandomFromRange)

		serviceGraphByte, err := yaml.Marshal(svcGraph)
		exitIfError(err)

		err = output.CreateAndPopulateFile(targetFilename, string(serviceGraphByte))
		exitIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(generateGraphCmd)
	generateGraphCmd.PersistentFlags().Int(
		"number-of-services", 0, "Number of service which will be created")
	generateGraphCmd.PersistentFlags().Int(
		"subtree-size", 5, "Max size of each subtree")
	generateGraphCmd.PersistentFlags().Int(
		"request-size", 10000, "Request size in bytes")
	generateGraphCmd.PersistentFlags().Int(
		"response-size", 100000, "Response size in bytes")
	generateGraphCmd.PersistentFlags().String(
		"cluster-list", "", "Comma separated list of cluster contexts")
	generateGraphCmd.PersistentFlags().String(
		"ingress-gateway-endpoint", "", "IP to ingress gateway")
}
