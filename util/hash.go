package util

import (
	"sort"

	"github.com/spaolacci/murmur3"
)

type Hrw struct {
	nodes map[string]int
}

func NewHrw() *Hrw {
	return &Hrw{
		nodes: make(map[string]int),
	}
}

func (hrw *Hrw) AddNode(nodes ...string) {
	for _, node := range nodes {
		hrw.nodes[node] = 1
	}
}

func (hrw *Hrw) RemoveNode(nodes ...string) {
	for _, node := range nodes {
		delete(hrw.nodes, node)
	}
}

type hrwSortNode struct {
	hash uint32
	val  string
}

func (hrw *Hrw) Get(value string) string {
	var values []hrwSortNode

	for node, _ := range hrw.nodes {
		hash := murmur3.Sum32([]byte(node + value))
		values = append(values, hrwSortNode{hash, node})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].hash > values[j].hash
	})

	return values[0].val
}
