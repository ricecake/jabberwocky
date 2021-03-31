package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"jabberwocky/util"
)

var _ = Describe("Util", func() {
	Context("HRW node selection", func(){
		It("Lives!", func(){
			util.NewHrw()
			Expect(1).To(Equal(1))
		})
	})
})
