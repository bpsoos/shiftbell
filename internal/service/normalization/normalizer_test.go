package normalization_test

import (
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
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
			SearchLimit:      3,
		})
	})

	Describe("NormalizeName", func() {
		It("normalizes whitespace and Unicode composition", func() {
			value, err := normalizer.NormalizeName("  e\u0301  ")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("é"))
		})

		It("accepts a name at the configured character limit", func() {
			value, err := normalizer.NormalizeName("abc")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("abc"))
		})

		It("counts multibyte UTF-8 values as single characters", func() {
			value, err := normalizer.NormalizeName("ééé")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("ééé"))
		})

		DescribeTable(
			"rejects an invalid name",
			func(value string, expected error) {
				normalized, err := normalizer.NormalizeName(value)

				Expect(err).To(MatchError(expected))
				Expect(normalized).To(BeEmpty())
			},
			Entry("when empty", "", validationerrors.ErrRequired),
			Entry("when whitespace only", " \t\n ", validationerrors.ErrRequired),
			Entry(
				"when invalid UTF-8",
				string([]byte{0xff}),
				validationerrors.ErrInvalidUTF8,
			),
			Entry("when over the configured limit", "abcd", validationerrors.ErrTooLong),
			Entry(
				"when containing a line break",
				"a\nb",
				validationerrors.ErrDisallowedCharacter,
			),
		)
	})

	Describe("NormalizeDescription", func() {
		It("normalizes whitespace, line endings, and Unicode composition", func() {
			value, err := normalizer.NormalizeDescription("  e\u0301\r\n\tx\r  ")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("é\n\tx"))
		})

		It("accepts an empty description", func() {
			value, err := normalizer.NormalizeDescription(" \t\r\n ")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(BeEmpty())
		})

		It("accepts a description at the configured character limit", func() {
			value, err := normalizer.NormalizeDescription("abcde")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("abcde"))
		})

		It("counts multibyte UTF-8 values as single characters", func() {
			value, err := normalizer.NormalizeDescription("ééééé")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("ééééé"))
		})

		DescribeTable(
			"rejects an invalid description",
			func(value string, expected error) {
				normalized, err := normalizer.NormalizeDescription(value)

				Expect(err).To(MatchError(expected))
				Expect(normalized).To(BeEmpty())
			},
			Entry(
				"when invalid UTF-8",
				string([]byte{0xff}),
				validationerrors.ErrInvalidUTF8,
			),
			Entry(
				"when over the configured limit",
				"abcdef",
				validationerrors.ErrTooLong,
			),
			Entry(
				"when containing a control character",
				"a\x00b",
				validationerrors.ErrDisallowedCharacter,
			),
		)
	})

	Describe("NormalizeSearch", func() {
		It("normalizes whitespace and Unicode composition", func() {
			value, err := normalizer.NormalizeSearch("  e\u0301  ")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("é"))
		})

		It("accepts an empty search", func() {
			value, err := normalizer.NormalizeSearch("   ")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(BeEmpty())
		})

		It("accepts a search at the configured character limit", func() {
			value, err := normalizer.NormalizeSearch("abc")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("abc"))
		})

		It("counts multibyte UTF-8 values as single characters", func() {
			value, err := normalizer.NormalizeSearch("ééé")

			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal("ééé"))
		})

		DescribeTable(
			"rejects an invalid search",
			func(value string, expected error) {
				normalized, err := normalizer.NormalizeSearch(value)

				Expect(err).To(MatchError(expected))
				Expect(normalized).To(BeEmpty())
			},
			Entry(
				"when invalid UTF-8",
				string([]byte{0xff}),
				validationerrors.ErrInvalidUTF8,
			),
			Entry("when over the configured limit", "abcd", validationerrors.ErrTooLong),
			Entry(
				"when containing a line break",
				"a\nb",
				validationerrors.ErrDisallowedCharacter,
			),
		)
	})
})
