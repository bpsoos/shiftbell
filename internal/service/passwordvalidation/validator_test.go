package passwordvalidation_test

import (
	"strings"

	"github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/bpsoos/shiftbell/internal/service/passwordvalidation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {
	var validator *passwordvalidation.Validator

	BeforeEach(func() {
		validator = passwordvalidation.New()
	})

	DescribeTable("accepts passwords without composition rules", func(password string) {
		Expect(validator.Validate(password)).To(Succeed())
	},
		Entry("exactly 15 characters", "abcdefghijklmno"),
		Entry("a passphrase", "a long passphrase"),
		Entry("multibyte Unicode", strings.Repeat("é", 15)),
		Entry("emoji", strings.Repeat("🔑", 15)),
		Entry("spaces count toward length", " abcdefghijklm "),
		Entry("whitespace only", strings.Repeat(" ", 15)),
		Entry("decomposed Unicode is not normalized", strings.Repeat("e\u0301", 8)),
		Entry("more than 64 characters", strings.Repeat("x", 128)),
	)

	DescribeTable(
		"explains invalid passwords without including their contents",
		func(password string, reason error, message string) {
			err := validator.Validate(password)

			Expect(err).To(MatchError(reason))
			Expect(err.Error()).To(Equal(message))
		},
		Entry("empty", "", validation.ErrRequired, "value is required"),
		Entry(
			"invalid UTF-8",
			strings.Repeat("a", 15)+string([]byte{0xff}),
			validation.ErrInvalidUTF8,
			"invalid UTF-8",
		),
		Entry(
			"14 ASCII characters",
			"abcdefghijklmn",
			validation.ErrTooShort,
			"value is too short: password must contain at least 15 characters; received 14",
		),
		Entry(
			"14 multibyte characters",
			strings.Repeat("é", 14),
			validation.ErrTooShort,
			"value is too short: password must contain at least 15 characters; received 14",
		),
		Entry(
			"14 emoji",
			strings.Repeat("🔑", 14),
			validation.ErrTooShort,
			"value is too short: password must contain at least 15 characters; received 14",
		),
	)
})
