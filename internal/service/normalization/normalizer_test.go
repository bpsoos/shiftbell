package normalization_test

import (
	"github.com/bpsoos/shiftbell/internal/service/normalization"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Normalizer", func() {
	var normalizer *normalization.Normalizer

	BeforeEach(func() {
		normalizer = normalization.New(normalization.Config{
			NameLimit:        3,
			DescriptionLimit: 5,
		})
	})

	Describe("NormalizeName", func() {
		It("normalizes whitespace and Unicode composition", func() {
			value, valid := normalizer.NormalizeName("  e\u0301  ")

			Expect(valid).To(BeTrue())
			Expect(value).To(Equal("é"))
		})

		It("accepts a name at the configured character limit", func() {
			value, valid := normalizer.NormalizeName("abc")

			Expect(valid).To(BeTrue())
			Expect(value).To(Equal("abc"))
		})

		It("counts multibyte UTF-8 values as single characters", func() {
			value, valid := normalizer.NormalizeName("ééé")

			Expect(valid).To(BeTrue())
			Expect(value).To(Equal("ééé"))
		})

		DescribeTable("rejects an invalid name",
			func(value string) {
				normalized, valid := normalizer.NormalizeName(value)

				Expect(valid).To(BeFalse())
				Expect(normalized).To(BeEmpty())
			},
			Entry("when empty", ""),
			Entry("when whitespace only", " \t\n "),
			Entry("when invalid UTF-8", string([]byte{0xff})),
			Entry("when over the configured limit", "abcd"),
			Entry("when containing a line break", "a\nb"),
		)
	})

	Describe("NormalizeDescription", func() {
		It("normalizes whitespace, line endings, and Unicode composition", func() {
			value, valid := normalizer.NormalizeDescription("  e\u0301\r\n\tx\r  ")

			Expect(valid).To(BeTrue())
			Expect(value).To(Equal("é\n\tx"))
		})

		It("accepts an empty description", func() {
			value, valid := normalizer.NormalizeDescription(" \t\r\n ")

			Expect(valid).To(BeTrue())
			Expect(value).To(BeEmpty())
		})

		It("accepts a description at the configured character limit", func() {
			value, valid := normalizer.NormalizeDescription("abcde")

			Expect(valid).To(BeTrue())
			Expect(value).To(Equal("abcde"))
		})

		It("counts multibyte UTF-8 values as single characters", func() {
			value, valid := normalizer.NormalizeDescription("ééééé")

			Expect(valid).To(BeTrue())
			Expect(value).To(Equal("ééééé"))
		})

		DescribeTable("rejects an invalid description",
			func(value string) {
				normalized, valid := normalizer.NormalizeDescription(value)

				Expect(valid).To(BeFalse())
				Expect(normalized).To(BeEmpty())
			},
			Entry("when invalid UTF-8", string([]byte{0xff})),
			Entry("when over the configured limit", "abcdef"),
			Entry("when containing a control character", "a\x00b"),
		)
	})
})
