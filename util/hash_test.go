package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"jabberwocky/util"
)

var _ = Describe("Util", func() {
	Context("HRW type", func() {
		var hrw *util.Hrw
		var nodes []string
		var entity string

		BeforeEach(func() {
			nodes = []string{
				"node_a", "node_b", "node_c",
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
				hrw.AddNode("node_d")

				By("checking that it's there")
				Expect(hrw.Nodes()).To(ContainElement("node_d"))
			})
			It("counts saved nodes", func() {
				By("checking preconditions")
				Expect(hrw.Size()).To(Equal(len(nodes)))

				By("adding a new node")
				hrw.AddNode("node_d")

				By("checking that it's there")
				Expect(hrw.Size()).To(Equal(len(nodes) + 1))
			})
			It("removes nodes", func() {
				Expect(hrw.Nodes()).To(ContainElement("node_a"))
				hrw.RemoveNode("node_a")
				Expect(hrw.Nodes()).ToNot(ContainElement("node_a"))
			})
			It("returns correct node", func() {
				selected := hrw.Get(entity)
				Expect(selected).To(Equal("node_b"))
			})

			It("returns correct node when an unrelated node removed", func() {
				hrw.RemoveNode("node_a")
				selected := hrw.Get(entity)
				Expect(selected).To(Equal("node_b"))
			})

			It("returns correct node when a related node removed", func() {
				hrw.RemoveNode("node_b")
				selected := hrw.Get(entity)
				Expect(selected).To(Equal("node_a"))
			})

		})
	})
})
