// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

//go:embed doc.tmpl
var docTemplate string

func main() {

	var clusterRoleFile string
	flag.StringVar(&clusterRoleFile, "file", "", "path to a file which contains a ClusterRole rules. Only one ClusterRole is supported.")
	flag.Parse()

	if clusterRoleFile == "" {
		panic(errors.New("missing required -file flag"))
	}

	rbac, err := os.ReadFile(clusterRoleFile)
	if err != nil {
		return
	}

	var node yaml.Node
	if err := yaml.Unmarshal(rbac, &node); err != nil {
		panic(err)
	}

	// Search for a ScalarNode with Value=rules
	var rules *yaml.Node
	nodes := node.Content[0]
	for i, node := range nodes.Content {
		if node.Value == "rules" {
			// Next node should hold the actual rules
			rules = nodes.Content[i+1]
			break
		}
	}

	documentedRules := make([]DocumentedRule, 0, len(rules.Content))
	// Process each rule
	for _, rule := range rules.Content {
		newRule := DocumentedRule{
			Comment: flattenComment(rule.HeadComment),
		}
		apiGroups, err := getSubNodesFor(rule.Content, "apiGroups")
		if err != nil {
			panic(err)
		}
		newRule.APIGroups = apiGroups

		resources, err := getSubNodesFor(rule.Content, "resources")
		if err != nil {
			panic(err)
		}
		newRule.Resources = resources

		verbs, err := getSubNodesFor(rule.Content, "verbs")
		if err != nil {
			panic(err)
		}
		newRule.Verbs = verbs

		nonResourceURLs, _ := getSubNodesFor(rule.Content, "nonResourceURLs")
		newRule.NonResourceURLs = nonResourceURLs

		resourceNames, _ := getSubNodesFor(rule.Content, "resourceNames")
		newRule.ResourceNames = resourceNames

		documentedRules = append(documentedRules, newRule)
	}

	tmpl, err := template.New("docTemplate").Parse(docTemplate)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(os.Stdout, documentedRules); err != nil {
		panic(err)
	}
}

// getSubNodesFor attempts to find the Yaml node with the provided name and returns its value if found.
func getSubNodesFor(nodes []*yaml.Node, nodeName string) ([]string, error) {
	apiGroups := sets.New[string]()
	for i, node := range nodes {
		if node.Value != nodeName {
			continue
		}
		// APIGroups should be right after...
		node := nodes[i+1]
		for _, apiGroup := range node.Content {
			if len(apiGroup.Value) == 0 {
				continue
			}
			apiGroups.Insert(apiGroup.Value)
		}
		return sets.List(apiGroups), nil
	}
	return nil, fmt.Errorf("node %s not found", nodeName)
}

func flattenComment(comment string) template.HTML {
	result := strings.ReplaceAll(comment, "#", "")
	result = strings.ReplaceAll(result, "\n", "")
	return template.HTML(strings.Join(strings.Fields(result), " "))

}

type DocumentedRule struct {
	APIGroups       []string
	Resources       []string
	Verbs           []string
	ResourceNames   []string
	NonResourceURLs []string
	Comment         template.HTML
}
