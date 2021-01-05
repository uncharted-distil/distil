//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package graph

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Edge represents a graph edge.
type Edge struct {
	Src        int                    `json:"source"`
	Target     int                    `json:"target"`
	Label      string                 `json:"label,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Node represents a graph node.
type Node struct {
	ID         int                    `json:"id"`
	Label      string                 `json:"label"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Graph represents a graph.
type Graph struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

var (
	graphRegex = regexp.MustCompile(`\s*graph\s*\[([\s\S]*)\]`)
	nodesRegex = regexp.MustCompile(`\s*node\s*\[([\s\S]*?)\]`)
	edgesRegex = regexp.MustCompile(`\s*edge\s*\[([\s\S]*?)\]`)
)

// NewNode instantiates a new node.
func NewNode() *Node {
	return &Node{
		Attributes: make(map[string]interface{}),
	}
}

// NewEdge instantiates a new edge.
func NewEdge() *Edge {
	return &Edge{
		Attributes: make(map[string]interface{}),
	}
}

func parseAmbiguous(arg string) interface{} {
	str, err := strconv.Unquote(arg)
	if err == nil {
		return str
	}

	i, err := strconv.ParseInt(arg, 10, 32)
	if err == nil {
		return i
	}

	f, err := strconv.ParseFloat(arg, 64)
	if err == nil {
		return f
	}

	b, err := strconv.ParseBool(arg)
	if err == nil {
		return b
	}

	return arg
}

func appendNodeField(n *Node, key string, value string) error {
	switch key {
	case "id":
		id, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "bad node format")
		}
		n.ID = int(id)
	case "label":
		label, err := strconv.Unquote(value)
		if err != nil {
			return errors.Wrap(err, "bad node label format")
		}
		n.Label = label
	default:

		n.Attributes[key] = parseAmbiguous(value)
	}
	return nil
}

func appendEdgeField(e *Edge, key string, value string) error {
	switch key {
	case "source":
		src, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		e.Src = int(src)
	case "target":
		target, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return errors.Wrap(err, "bad node id format")
		}
		e.Target = int(target)
	case "label":
		label, err := strconv.Unquote(value)
		if err != nil {
			return errors.Wrap(err, "bad label format")
		}
		e.Label = label
	default:

		e.Attributes[key] = parseAmbiguous(value)
	}
	return nil
}

// ParseGML parses a GML file into a set of graphs.
func ParseGML(gml string) ([]*Graph, error) {

	graphMatches := graphRegex.FindAllStringSubmatch(gml, -1)

	if len(graphMatches) == 0 {
		return nil, fmt.Errorf("no graph found")
	}

	var graphs []*Graph

	for _, graph := range graphMatches {

		if len(graph) != 2 {
			return nil, fmt.Errorf("error parsing graph")
		}
		graphInternal := graph[1]

		g := &Graph{}

		nodeMatches := nodesRegex.FindAllStringSubmatch(graphInternal, -1)

		if len(nodeMatches) == 0 {
			return nil, fmt.Errorf("no nodes found in graph")
		}

		for _, node := range nodeMatches {

			if len(node) != 2 {
				return nil, fmt.Errorf("error parsing node")
			}
			nodeInternal := node[1]

			n := NewNode()
			lines := strings.Split(nodeInternal, "\n")
			for _, line := range lines {

				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				if len(fields) != 2 {
					return nil, fmt.Errorf("error parsing fields in line: `%s`", line)
				}
				key := fields[0]
				value := fields[1]
				if err := appendNodeField(n, key, value); err != nil {
					return nil, err
				}
			}
			g.Nodes = append(g.Nodes, n)
		}

		edgeMatches := edgesRegex.FindAllStringSubmatch(graphInternal, -1)

		if len(edgeMatches) == 0 {
			return nil, fmt.Errorf("no nodes found in graph")
		}

		for _, edge := range edgeMatches {

			if len(edge) != 2 {
				return nil, fmt.Errorf("error parsing edge")
			}
			edgeInternal := edge[1]

			e := NewEdge()
			lines := strings.Split(edgeInternal, "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				if len(fields) != 2 {
					return nil, fmt.Errorf("error parsing fields in line: `%s`", line)
				}
				key := fields[0]
				value := fields[1]
				if err := appendEdgeField(e, key, value); err != nil {
					return nil, err
				}
			}

			g.Edges = append(g.Edges, e)
		}

		graphs = append(graphs, g)
	}
	return graphs, nil
}
