package usernamevalidation_test

import (
	"strings"

	"github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/bpsoos/shiftbell/internal/service/usernamevalidation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {
	var validator *usernamevalidation.Validator

	BeforeEach(func() {
		validator = usernamevalidation.New(usernamevalidation.Config{MaxLength: 12})
	})

	DescribeTable(
		"returns the canonical username",
		func(username, expected string) {
			canonical, err := validator.Validate(username)

			Expect(err).NotTo(HaveOccurred())
			Expect(canonical).To(Equal(expected))
		},
		Entry("mixed case", "Owner", "Owner"),
		Entry("composed Unicode", "Álex", "Álex"),
		Entry("decomposed Unicode", "A\u0301lex", "Álex"),
		Entry("punctuation", "O'Connor", "O'Connor"),
		Entry("at the length limit", "abcdefghijkl", "abcdefghijkl"),
		Entry(
			"multibyte characters at the length limit",
			strings.Repeat("é", 12),
			strings.Repeat("é", 12),
		),
		Entry(
			"decomposed characters at the length limit",
			strings.Repeat("e\u0301", 12),
			strings.Repeat("é", 12),
		),
	)

	DescribeTable(
		"explains invalid input",
		func(username string, reason error, message string) {
			canonical, err := validator.Validate(username)

			Expect(err).To(MatchError(reason))
			Expect(err.Error()).To(Equal(message))
			Expect(canonical).To(BeEmpty())
		},
		Entry("empty", "", validation.ErrRequired, "value is required"),
		Entry(
			"invalid UTF-8",
			string([]byte{0xff}),
			validation.ErrInvalidUTF8,
			"invalid UTF-8",
		),
		Entry(
			"leading space",
			" Owner",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace ' ' (U+0020) at character 1 is not allowed",
		),
		Entry(
			"trailing space",
			"Owner ",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace ' ' (U+0020) at character 6 is not allowed",
		),
		Entry(
			"internal space after multibyte text",
			"Álex Smith",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace ' ' (U+0020) at character 5 is not allowed",
		),
		Entry(
			"tab",
			"Ow\tner",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace '\\t' (U+0009) at character 3 is not allowed",
		),
		Entry(
			"newline",
			"Owner\n",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace '\\n' (U+000A) at character 6 is not allowed",
		),
		Entry(
			"nonbreaking space",
			"Ow\u00a0ner",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace '\\u00a0' (U+00A0) at character 3 is not allowed",
		),
		Entry(
			"Unicode em space",
			"Ow\u2003ner",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace '\\u2003' (U+2003) at character 3 is not allowed",
		),
		Entry(
			"whitespace only",
			" ",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace ' ' (U+0020) at character 1 is not allowed",
		),
		Entry(
			"control character",
			"Ow\x00ner",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: non-printable character '\\x00' (U+0000) at character 3 is not allowed",
		),
		Entry(
			"zero-width space",
			"Ow\u200bner",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: non-printable character '\\u200b' (U+200B) at character 3 is not allowed",
		),
		Entry("too long", "abcdefghijklm", validation.ErrTooLong,
			"value is too long: maximum is 12 characters; received 13"),
		Entry(
			"too long after NFC normalization",
			strings.Repeat("e\u0301", 13),
			validation.ErrTooLong,
			"value is too long: maximum is 12 characters; received 13",
		),
		Entry(
			"whitespace position in original decomposed input",
			"A\u0301lex ",
			validation.ErrDisallowedCharacter,
			"value contains a disallowed character: whitespace ' ' (U+0020) at character 6 is not allowed",
		),
	)
})
