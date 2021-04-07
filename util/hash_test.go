package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"jabberwocky/util"
)

type testNode struct {
	name string
}

func (tn testNode) HashKey() string {
	return tn.name
}
func (tn testNode) HashWeight() int {
	return 1
}

func mkTn(s string) testNode {
	return testNode{s}
}

var _ = Describe("Util", func() {
	Context("HRW type", func() {
		var hrw *util.Hrw
		var nodes []util.HrwNode
		var entity string

		BeforeEach(func() {
			nodes = []util.HrwNode{
				mkTn("node_a"), mkTn("node_b"), mkTn("node_c"),
			}
			entity = "someAgent"
		})

		JustBeforeEach(func() {
			hrw = util.NewHrw()
			hrw.AddNode(nodes...)
		})

		Describe("Basic functionality", func() {
			It("stores saved nodes", func() {
				By("checking preconditions")
				for _, node := range nodes {
					Expect(hrw.Nodes()).To(ContainElement(node))
				}

				By("adding a new node")
				hrw.AddNode(mkTn("node_d"))

				By("checking that it's there")
				Expect(hrw.Nodes()).To(ContainElement(mkTn("node_d")))
			})
			It("counts saved nodes", func() {
				By("checking preconditions")
				Expect(hrw.Size()).To(Equal(len(nodes)))

				By("adding a new node")
				hrw.AddNode(mkTn("node_d"))

				By("checking that it's there")
				Expect(hrw.Size()).To(Equal(len(nodes) + 1))
			})
			It("removes nodes", func() {
				Expect(hrw.Nodes()).To(ContainElement(mkTn("node_a")))
				hrw.RemoveNode(mkTn("node_a"))
				Expect(hrw.Nodes()).ToNot(ContainElement(mkTn("node_a")))
			})
			It("returns correct node", func() {
				selected := hrw.Get(entity)
				Expect(selected).To(Equal(mkTn("node_b")))
			})

			It("returns correct node when an unrelated node removed", func() {
				hrw.RemoveNode(mkTn("node_a"))
				selected := hrw.Get(entity)
				Expect(selected).To(Equal(mkTn("node_b")))
			})

			It("returns correct node when a related node removed", func() {
				hrw.RemoveNode(mkTn("node_b"))
				selected := hrw.Get(entity)
				Expect(selected).To(Equal(mkTn("node_a")))
			})

		})
	})
})
