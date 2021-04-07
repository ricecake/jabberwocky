package util

import (
	"sort"

	"github.com/spaolacci/murmur3"
)

type HrwNode interface {
	HashKey() string
	HashWeight() int
}

type Hrw struct {
	nodes map[string]HrwNode
}

func NewHrw() *Hrw {
	return &Hrw{
		nodes: make(map[string]HrwNode),
	}
}

func (hrw *Hrw) AddNode(nodes ...HrwNode) {
	for _, node := range nodes {
		val := node.HashKey()
		hrw.nodes[val] = node
	}
}

func (hrw *Hrw) RemoveNode(nodes ...HrwNode) {
	for _, node := range nodes {
		delete(hrw.nodes, node.HashKey())
	}
}

func (hrw *Hrw) Size() int {
	return len(hrw.nodes)
}

func (hrw *Hrw) Nodes() (nodes []HrwNode) {
	for _, node := range hrw.nodes {
		nodes = append(nodes, node)
	}
	return
}

type hrwSortNode struct {
	hash uint32
	val  HrwNode
}

func (hrw *Hrw) Get(value string) HrwNode {
	var values []hrwSortNode

	for key, node := range hrw.nodes {
		hash := murmur3.Sum32([]byte(key + value))
		values = append(values, hrwSortNode{hash, node})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].hash > values[j].hash
	})

	return values[0].val
}
