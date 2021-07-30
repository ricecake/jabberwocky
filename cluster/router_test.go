package cluster_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"jabberwocky/cluster"
)

var _ = Describe("Router", func() {
	/*
		Context("forwarding choices", func() {
			BeforeEach(func() {
			})

			JustBeforeEach(func() {
			})

			Describe("Basic functionality", func() {
				It("stores saved nodes", func() {
					By("checking preconditions")
				})
			})
		})
		Context("Superset router", func() {
			BeforeEach(func() {
			})

			JustBeforeEach(func() {
			})

			Describe("Basic functionality", func() {
				It("stores saved nodes", func() {
					By("checking preconditions")
				})
			})
		})
	*/
	Context("Subset router", func() {
		var subRouter *cluster.SubsetRouter
		var dests []cluster.Destination

		JustBeforeEach(func() {
			subRouter = cluster.NewSubsetRouter()
		})

		Describe("Basic functionality", func() {
			BeforeEach(func() {
				for _, v := range []string{"a", "b", "c", "d"} {
					dests = append(dests, cluster.Destination{
						Role: cluster.LOCAL_AGENT,
						Code: v,
					})
				}
			})
			It("handles catch-all with emtpy tags", func() {
				By("Adding empty route")
				subRouter.AddBind(dests[0], map[string]string{})

				By("Getting routes for empty tags")
				routes := subRouter.Route(map[string]string{})

				Expect(routes).Should(ConsistOf(dests[0]))
			})

			It("handles catch-all with non-empty tags", func() {
				By("Adding empty route")
				subRouter.AddBind(dests[0], map[string]string{})

				By("Getting routes for non-empty tags")
				routes := subRouter.Route(map[string]string{"a": "b"})

				Expect(routes).Should(ConsistOf(dests[0]))
			})
			It("handles exact match", func() {
				subRouter.AddBind(dests[0], map[string]string{"a": "b"})
				routes := subRouter.Route(map[string]string{"a": "b"})
				Expect(routes).Should(ConsistOf(dests[0]))
			})

			It("handles inexact match", func() {
				subRouter.AddBind(dests[0], map[string]string{"a": "b"})
				routes := subRouter.Route(map[string]string{"a": "b", "c": "d"})
				Expect(routes).Should(ConsistOf(dests[0]))
			})
			It("handles no overlap", func() {
				subRouter.AddBind(dests[0], map[string]string{"a": "b"})
				routes := subRouter.Route(map[string]string{"c": "d"})
				Expect(routes).Should(BeEmpty())
			})
			It("handles partial overlap", func() {
				subRouter.AddBind(dests[0], map[string]string{"a": "b", "c": "d"})
				routes := subRouter.Route(map[string]string{"a": "b", "e": "f"})
				Expect(routes).Should(BeEmpty())
			})
		})
	})
})
