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
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"html/template"
	"os"
	"sort"
	"strings"
)

//go:embed doc.tmpl
var docTemplate string

const (
	StartMarker = "<!--- START CLUSTER ROLES DOCUMENTATION --->"
	EndMarker   = "<!--- END CLUSTER ROLES DOCUMENTATION --->"
)

func createTableHeader(hasNonResourceURLS, hasResourceNames bool) string {
	tableHeader := "| API Groups  | Resources |"
	tableHeaderSeparator := "| ----------- | --------- |"
	if hasNonResourceURLS {
		tableHeader = tableHeader + " Non Resource URLs |"
		tableHeaderSeparator = tableHeaderSeparator + " ----- |"
	}
	if hasResourceNames {
		tableHeader = tableHeader + " Resource Names |"
		tableHeaderSeparator = tableHeaderSeparator + " -------------- |"
	}
	tableHeader = tableHeader + " Verbs | Comment |"
	tableHeaderSeparator = tableHeaderSeparator + " ----------------- | ----------------- |"
	return tableHeader + "\n" + tableHeaderSeparator
}

func main() {

	var clusterRoleFile, out string
	flag.StringVar(&clusterRoleFile, "in", "", "path to a file which contains a ClusterRole rules. Only one ClusterRole is supported.")
	flag.StringVar(&out, "out", "", "path to a file were the content should be written")
	flag.Parse()

	if clusterRoleFile == "" {
		panic(errors.New("missing required -file flag"))
	}

	rbac, err := os.ReadFile(clusterRoleFile)
	if err != nil {
		panic(err)
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

	documentedRuleList := make([]DocumentedRule, 0, len(rules.Content))
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

		documentedRuleList = append(documentedRuleList, newRule)
	}

	tmpl, err := template.New("docTemplate").Parse(docTemplate)
	if err != nil {
		panic(err)
	}

	// If out is not nil:
	// 1. Check if output file exist
	// 2. If not, create it with the marker
	// 3. If it does exist, read until the marker
	// 3.1 If not marker found append content to the end of file
	// 4. Write the output
	// 5. Skip content until end marker

	documentedRules := NewDocumentedRules(documentedRuleList)
	buff := new(bytes.Buffer)
	if err := tmpl.Execute(buff, documentedRules); err != nil {
		panic(err)
	}

	if out == "" {
		fmt.Println(string(buff.Bytes()))
		return
	}

	// Check if file does exist.
	if _, err := os.Stat(out); errors.Is(err, os.ErrNotExist) {
		// No, let's just create it.
		if err := os.WriteFile(out, buff.Bytes(), 0644); err != nil {
			panic(err)
		}
		return
	}

	// Yes, insert or append the content.
	updatedContent := updateMarkers(out, StartMarker, EndMarker, buff.Bytes())
	if err := os.WriteFile(out, updatedContent, 0644); err != nil {
		panic(err)
	}
}

// getSubNodesFor attempts to find the Yaml node with the provided name and returns its value if found.
func getSubNodesFor(nodes []*yaml.Node, nodeName string) ([]string, error) {
	apiGroups := make(map[string]struct{})
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
			apiGroups[apiGroup.Value] = struct{}{}
		}
		return toList(apiGroups), nil
	}
	return nil, fmt.Errorf("node %s not found", nodeName)
}

func toList(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func flattenComment(comment string) template.HTML {
	result := strings.ReplaceAll(comment, "#", "")
	result = strings.ReplaceAll(result, "\n", "")
	return template.HTML(strings.Join(strings.Fields(result), " "))

}

func NewDocumentedRules(documentedRules DocumentedRuleList) DocumentedRules {
	return DocumentedRules{
		TableHeader: createTableHeader(
			documentedRules.HasResourceNames(),
			documentedRules.HasNonResourceURLs(),
		),
		DocumentedRules: documentedRules,
	}
}

type DocumentedRules struct {
	TableHeader     string
	DocumentedRules []DocumentedRule
}

type DocumentedRuleList []DocumentedRule

func (dl *DocumentedRuleList) HasResourceNames() bool {
	if dl == nil {
		return false
	}
	for _, rule := range *dl {
		if len(rule.ResourceNames) > 0 {
			return true
		}
	}
	return false
}

func (dl *DocumentedRuleList) HasNonResourceURLs() bool {
	if dl == nil {
		return false
	}
	for _, rule := range *dl {
		if len(rule.NonResourceURLs) > 0 {
			return true
		}
	}
	return false
}

type DocumentedRule struct {
	APIGroups       []string
	Resources       []string
	Verbs           []string
	ResourceNames   []string
	NonResourceURLs []string
	Comment         template.HTML
}
