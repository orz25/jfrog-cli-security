package sca

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"reflect"
	"testing"

	"golang.org/x/exp/maps"

	"github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	coreXray "github.com/jfrog/jfrog-cli-core/v2/utils/xray"
	"github.com/jfrog/jfrog-cli-security/utils"
	"github.com/jfrog/jfrog-client-go/xray/services"
	xrayUtils "github.com/jfrog/jfrog-client-go/xray/services/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetExcludePattern(t *testing.T) {
	tests := []struct {
		name     string
		params   func() *utils.AuditBasicParams
		expected string
	}{
		{
			name: "Test exclude pattern recursive",
			params: func() *utils.AuditBasicParams {
				param := &utils.AuditBasicParams{}
				param.SetExclusions([]string{"exclude1", "exclude2"}).SetIsRecursiveScan(true)
				return param
			},
			expected: "(^exclude1$)|(^exclude2$)",
		},
		{
			name:     "Test no exclude pattern recursive",
			params:   func() *utils.AuditBasicParams { return (&utils.AuditBasicParams{}).SetIsRecursiveScan(true) },
			expected: "(^.*\\.git.*$)|(^.*node_modules.*$)|(^.*target.*$)|(^.*venv.*$)|(^.*test.*$)",
		},
		{
			name: "Test exclude pattern not recursive",
			params: func() *utils.AuditBasicParams {
				param := &utils.AuditBasicParams{}
				param.SetExclusions([]string{"exclude1", "exclude2"})
				return param
			},
			expected: "(^exclude1$)|(^exclude2$)",
		},
		{
			name:     "Test no exclude pattern",
			params:   func() *utils.AuditBasicParams { return &utils.AuditBasicParams{} },
			expected: "(^.*\\.git.*$)|(^.*node_modules.*$)|(^.*target.*$)|(^.*venv.*$)|(^.*test.*$)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetExcludePattern(test.params())
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestBuildXrayDependencyTree(t *testing.T) {
	treeHelper := make(map[string]coreXray.DepTreeNode)
	rootDep := coreXray.DepTreeNode{Children: []string{"topDep1", "topDep2", "topDep3"}}
	topDep1 := coreXray.DepTreeNode{Children: []string{"midDep1", "midDep2"}}
	topDep2 := coreXray.DepTreeNode{Children: []string{"midDep2", "midDep3"}}
	midDep1 := coreXray.DepTreeNode{Children: []string{"bottomDep1"}}
	midDep2 := coreXray.DepTreeNode{Children: []string{"bottomDep2", "bottomDep3"}}
	bottomDep3 := coreXray.DepTreeNode{Children: []string{"leafDep"}}
	treeHelper["rootDep"] = rootDep
	treeHelper["topDep1"] = topDep1
	treeHelper["topDep2"] = topDep2
	treeHelper["midDep1"] = midDep1
	treeHelper["midDep2"] = midDep2
	treeHelper["bottomDep3"] = bottomDep3

	expectedUniqueDeps := []string{"rootDep", "topDep1", "topDep2", "topDep3", "midDep1", "midDep2", "midDep3", "bottomDep1", "bottomDep2", "bottomDep3", "leafDep"}

	// Constructing the expected tree Nodes
	leafDepNode := &xrayUtils.GraphNode{Id: "leafDep", Nodes: []*xrayUtils.GraphNode{}}
	bottomDep3Node := &xrayUtils.GraphNode{Id: "bottomDep3", Nodes: []*xrayUtils.GraphNode{}}
	bottomDep2Node := &xrayUtils.GraphNode{Id: "bottomDep2", Nodes: []*xrayUtils.GraphNode{}}
	bottomDep1Node := &xrayUtils.GraphNode{Id: "bottomDep1", Nodes: []*xrayUtils.GraphNode{}}
	midDep3Node := &xrayUtils.GraphNode{Id: "midDep3", Nodes: []*xrayUtils.GraphNode{}}
	midDep2Node := &xrayUtils.GraphNode{Id: "midDep2", Nodes: []*xrayUtils.GraphNode{}}
	midDep1Node := &xrayUtils.GraphNode{Id: "midDep1", Nodes: []*xrayUtils.GraphNode{}}
	topDep3Node := &xrayUtils.GraphNode{Id: "topDep3", Nodes: []*xrayUtils.GraphNode{}}
	topDep2Node := &xrayUtils.GraphNode{Id: "topDep2", Nodes: []*xrayUtils.GraphNode{}}
	topDep1Node := &xrayUtils.GraphNode{Id: "topDep1", Nodes: []*xrayUtils.GraphNode{}}
	rootNode := &xrayUtils.GraphNode{Id: "rootDep", Nodes: []*xrayUtils.GraphNode{}}

	// Setting children to parents
	bottomDep3Node.Nodes = append(bottomDep3Node.Nodes, leafDepNode)
	midDep2Node.Nodes = append(midDep2Node.Nodes, bottomDep3Node)
	midDep2Node.Nodes = append(midDep2Node.Nodes, bottomDep2Node)
	midDep1Node.Nodes = append(midDep1Node.Nodes, bottomDep1Node)
	topDep2Node.Nodes = append(topDep2Node.Nodes, midDep3Node)
	topDep2Node.Nodes = append(topDep2Node.Nodes, midDep2Node)
	topDep1Node.Nodes = append(topDep1Node.Nodes, midDep2Node)
	topDep1Node.Nodes = append(topDep1Node.Nodes, midDep1Node)
	rootNode.Nodes = append(rootNode.Nodes, topDep1Node)
	rootNode.Nodes = append(rootNode.Nodes, topDep2Node)
	rootNode.Nodes = append(rootNode.Nodes, topDep3Node)

	// Setting children to parents
	leafDepNode.Parent = bottomDep3Node
	bottomDep3Node.Parent = midDep2Node
	bottomDep3Node.Parent = midDep2Node
	bottomDep1Node.Parent = midDep1Node
	midDep3Node.Parent = topDep2Node
	midDep2Node.Parent = topDep2Node
	midDep2Node.Parent = topDep1Node
	midDep1Node.Parent = topDep1Node
	topDep1Node.Parent = rootNode
	topDep2Node.Parent = rootNode
	topDep3Node.Parent = rootNode

	tree, uniqueDeps := coreXray.BuildXrayDependencyTree(treeHelper, "rootDep")

	assert.ElementsMatch(t, expectedUniqueDeps, maps.Keys(uniqueDeps))
	assert.True(t, tests.CompareTree(tree, rootNode))
}

func TestSetPathsForIssues(t *testing.T) {
	// Create a test dependency tree
	rootNode := &xrayUtils.GraphNode{Id: "root"}
	childNode1 := &xrayUtils.GraphNode{Id: "child1"}
	childNode2 := &xrayUtils.GraphNode{Id: "child2"}
	childNode3 := &xrayUtils.GraphNode{Id: "child3"}
	childNode4 := &xrayUtils.GraphNode{Id: "child4"}
	childNode5 := &xrayUtils.GraphNode{Id: "child5"}
	rootNode.Nodes = []*xrayUtils.GraphNode{childNode1, childNode2, childNode3}
	childNode2.Nodes = []*xrayUtils.GraphNode{childNode4}
	childNode3.Nodes = []*xrayUtils.GraphNode{childNode5}

	// Create a test issues map
	issuesMap := make(map[string][][]services.ImpactPathNode)
	issuesMap["child1"] = [][]services.ImpactPathNode{}
	issuesMap["child4"] = [][]services.ImpactPathNode{}
	issuesMap["child5"] = [][]services.ImpactPathNode{}

	// Call setPathsForIssues with the test data
	setPathsForIssues(rootNode, issuesMap, []services.ImpactPathNode{})

	// Check the results
	assert.Equal(t, issuesMap["child1"][0][0].ComponentId, "root")
	assert.Equal(t, issuesMap["child1"][0][1].ComponentId, "child1")

	assert.Equal(t, issuesMap["child4"][0][0].ComponentId, "root")
	assert.Equal(t, issuesMap["child4"][0][1].ComponentId, "child2")
	assert.Equal(t, issuesMap["child4"][0][2].ComponentId, "child4")

	assert.Equal(t, issuesMap["child5"][0][0].ComponentId, "root")
	assert.Equal(t, issuesMap["child5"][0][1].ComponentId, "child3")
	assert.Equal(t, issuesMap["child5"][0][2].ComponentId, "child5")
}

func TestUpdateVulnerableComponent(t *testing.T) {
	components := map[string]services.Component{
		"dependency1": {
			FixedVersions: []string{"1.0.0"},
			ImpactPaths:   [][]services.ImpactPathNode{},
		},
	}
	dependencyName, issuesMap := "dependency1", map[string][][]services.ImpactPathNode{
		"dependency1": {},
	}

	updateComponentsWithImpactPaths(components, issuesMap)

	// Check the result
	expected := services.Component{
		FixedVersions: []string{"1.0.0"},
		ImpactPaths:   issuesMap[dependencyName],
	}
	assert.Equal(t, expected, components[dependencyName])
}

func TestBuildImpactPaths(t *testing.T) {
	// create sample scan result and dependency trees
	scanResult := []services.ScanResponse{
		{
			Vulnerabilities: []services.Vulnerability{
				{
					Components: map[string]services.Component{
						"dep1": {
							FixedVersions: []string{"1.2.3"},
							Cpes:          []string{"cpe:/o:vendor:product:1.2.3"},
						},
						"dep2": {
							FixedVersions: []string{"3.0.0"},
						},
					},
				},
			},
			Violations: []services.Violation{
				{
					Components: map[string]services.Component{
						"dep2": {
							FixedVersions: []string{"4.5.6"},
							Cpes:          []string{"cpe:/o:vendor:product:4.5.6"},
						},
					},
				},
			},
			Licenses: []services.License{
				{
					Components: map[string]services.Component{
						"dep3": {
							FixedVersions: []string{"7.8.9"},
							Cpes:          []string{"cpe:/o:vendor:product:7.8.9"},
						},
					},
				},
			},
		},
	}
	dependencyTrees := []*xrayUtils.GraphNode{
		{
			Id: "dep1",
			Nodes: []*xrayUtils.GraphNode{
				{
					Id: "dep2",
					Nodes: []*xrayUtils.GraphNode{
						{
							Id:    "dep3",
							Nodes: []*xrayUtils.GraphNode{},
						},
					},
				},
			},
		},
		{
			Id: "dep7",
			Nodes: []*xrayUtils.GraphNode{
				{
					Id: "dep4",
					Nodes: []*xrayUtils.GraphNode{
						{
							Id:    "dep2",
							Nodes: []*xrayUtils.GraphNode{},
						},
						{
							Id:    "dep5",
							Nodes: []*xrayUtils.GraphNode{},
						},
						{
							Id:    "dep6",
							Nodes: []*xrayUtils.GraphNode{},
						},
					},
				},
			},
		},
	}

	scanResult = BuildImpactPathsForScanResponse(scanResult, dependencyTrees)
	// assert that the components were updated with impact paths
	expectedImpactPaths := [][]services.ImpactPathNode{{{ComponentId: "dep1"}}}
	assert.Equal(t, expectedImpactPaths, scanResult[0].Vulnerabilities[0].Components["dep1"].ImpactPaths)
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep1"}, {ComponentId: "dep2"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Vulnerabilities[0].Components["dep2"].ImpactPaths[0])
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep7"}, {ComponentId: "dep4"}, {ComponentId: "dep2"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Vulnerabilities[0].Components["dep2"].ImpactPaths[1])
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep1"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Violations[0].Components["dep1"].ImpactPaths)
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep1"}, {ComponentId: "dep2"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Violations[0].Components["dep2"].ImpactPaths[0])
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep7"}, {ComponentId: "dep4"}, {ComponentId: "dep2"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Violations[0].Components["dep2"].ImpactPaths[1])
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep7"}, {ComponentId: "dep4"}, {ComponentId: "dep2"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Violations[0].Components["dep2"].ImpactPaths)
	expectedImpactPaths = [][]services.ImpactPathNode{{{ComponentId: "dep1"}, {ComponentId: "dep2"}, {ComponentId: "dep3"}}}
	reflect.DeepEqual(expectedImpactPaths, scanResult[0].Licenses[0].Components["dep3"].ImpactPaths)
}

func TestSuspectCurationBlockedError(t *testing.T) {
	mvnOutput1 := "status code: 403, reason phrase: Forbidden (403)"
	mvnOutput2 := "status code: 500, reason phrase: Server Error (500)"
	pipOutput := "because of HTTP error 403 Client Error: Forbidden for url"

	tests := []struct {
		name          string
		isCurationCmd bool
		tech          coreutils.Technology
		output        string
		expect        string
	}{
		{
			name:          "mvn 403 error",
			isCurationCmd: true,
			tech:          coreutils.Maven,
			output:        mvnOutput1,
			expect:        fmt.Sprintf(curationErrorMsgToUserTemplate, coreutils.Maven),
		},
		{
			name:          "mvn 500 error",
			isCurationCmd: true,
			tech:          coreutils.Maven,
			output:        mvnOutput2,
			expect:        fmt.Sprintf(curationErrorMsgToUserTemplate, coreutils.Maven),
		},
		{
			name:          "pip 403 error",
			isCurationCmd: true,
			tech:          coreutils.Maven,
			output:        pipOutput,
			expect:        fmt.Sprintf(curationErrorMsgToUserTemplate, coreutils.Pip),
		},
		{
			name:          "pip not pass through error",
			isCurationCmd: true,
			tech:          coreutils.Pip,
			output:        "http error 401",
		},
		{
			name:          "maven not pass through error",
			isCurationCmd: true,
			tech:          coreutils.Maven,
			output:        "http error 401",
		},
		{
			name:          "nota supported tech",
			isCurationCmd: true,
			tech:          coreutils.CI,
			output:        pipOutput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SuspectCurationBlockedError(tt.isCurationCmd, tt.tech, tt.output)
		})
	}
}
