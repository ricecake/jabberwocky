package cluster

import (
	"github.com/davecgh/go-spew/spew"
)

type SubsetRouter struct {
	root *subsetRouteNode
}

type subsetRouteNode struct {
	bind        bindEntry
	subscribers []Destination
	subnode     map[string]map[string]*subsetRouteNode
}

func NewSubsetRouter() *SubsetRouter {
	return &SubsetRouter{
		root: &subsetRouteNode{
			subnode: make(map[string]map[string]*subsetRouteNode),
		},
	}
}

func (sr *SubsetRouter) Dump() {
	spew.Dump(sr)
}

// TODO
// Need a method to replace bindings by tag.  This will be used when merging remote state.
// Should be able to add all tags from new set, and then loop through and remove any bindings not in that new set.
// When gossiping state, servers should share what messages they are interested in being notified about.

func (sr *SubsetRouter) ReplaceBind(dest Destination, bindingis []map[string]string) {
	Given new binding left, and old binding right
	linearize left --- find a way to hash a binding?
	linearize right

	calculate intersection, and subtract from each.
	add left-prime
	remove right-prime
}

func (sr *SubsetRouter) AddBind(dest Destination, binding map[string]string) {
	if len(binding) == 0 {
		sr.root.subscribers = append(sr.root.subscribers, dest)
		return
	}

	bindings := linearizeTags(binding)

	node := sr.root
	for _, v := range bindings {
		var found bool
		var valMap map[string]*subsetRouteNode
		if valMap, found = node.subnode[v.key]; !found {
			valMap = make(map[string]*subsetRouteNode)
			node.subnode[v.key] = valMap
		}

		var subnode *subsetRouteNode
		if subnode, found = valMap[v.value]; !found {
			subnode = &subsetRouteNode{
				subnode: make(map[string]map[string]*subsetRouteNode),
				bind:    bindEntry{key: v.key, value: v.value},
			}
			valMap[v.value] = subnode
		}

		node = subnode
	}

	node.subscribers = append(node.subscribers, dest)
	// TODO this should ensure that there's only once copy of dest in the array.
}

func (sr *SubsetRouter) Route(tags map[string]string) (destinations []Destination) {
	if len(tags) == 0 {
		destinations = append(destinations, sr.root.subscribers...)
		return
	}

	type searchPath struct {
		node  *subsetRouteNode
		check []bindEntry
	}

	bindings := linearizeTags(tags)
	searchSpace := []searchPath{searchPath{sr.root, bindings}}

	for len(searchSpace) != 0 {
		item := searchSpace[0]
		searchSpace = searchSpace[1:]

		destinations = append(destinations, item.node.subscribers...)
		for index, bind := range item.check {
			if valMap, found := item.node.subnode[bind.key]; found {
				if subnode, found := valMap[bind.value]; found {
					searchSpace = append(searchSpace, searchPath{subnode, item.check[index+1:]})
				}
			}
		}
	}

	//TODO this should deduplicate subscribers.

	return
}

func (sr *SubsetRouter) ListLocalBinding() (output []map[string]string) {
	type searchNode struct {
		node  *subsetRouteNode
		state map[string]string
	}

	searchSpace := []searchNode{searchNode{sr.root, make(map[string]string)}}

	for len(searchSpace) != 0 {
		item := searchSpace[0]
		searchSpace = searchSpace[1:]

		for _, subscriber := range item.node.subscribers {
			if subscriber.Role.IsLocal() {
				output = append(output, copyTagMap(item.state))
				break
			}
		}

		for key, valueMap := range item.node.subnode {
			for value, subnode := range valueMap {
				state := copyTagMap(item.state)
				state[key] = value
				searchSpace = append(searchSpace, searchNode{subnode, state})
			}
		}
	}

	return
}
