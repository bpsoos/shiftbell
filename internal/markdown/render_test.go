package markdown_test

import (
	"testing"

	"github.com/bpsoos/shiftbell/internal/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMarkdown(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Markdown Renderer Suite")
}

var _ = Describe("Render", func() {
	It("renders standard Markdown", func() {
		output, err := markdown.Render(
			"# Instructions\n\n**Clean** the *kitchen*.\n\n- Counters\n- Sink\n\n[Guide](https://example.com)\n\n> Carefully\n\n```\n<sample>\n```",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("<h1>Instructions</h1>"))
		Expect(output).To(ContainSubstring("<strong>Clean</strong>"))
		Expect(output).To(ContainSubstring("<em>kitchen</em>"))
		Expect(output).To(ContainSubstring("<li>Counters</li>"))
		Expect(output).To(ContainSubstring(`<a href="https://example.com">Guide</a>`))
		Expect(output).To(ContainSubstring("<blockquote>"))
		Expect(output).To(ContainSubstring("<pre><code>&lt;sample&gt;"))
	})

	It("preserves single line breaks and separate paragraphs", func() {
		output, err := markdown.Render("First\nSecond\n\nThird")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal("<p>First<br>\nSecond</p>\n<p>Third</p>\n"))
	})

	It("omits raw HTML and unsafe link destinations", func() {
		output, err := markdown.Render(
			"<script>alert(1)</script>\n\n[Click](javascript:alert%281%29)\n\n[Data](data:text/html,test)",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).NotTo(ContainSubstring("<script>"))
		Expect(output).NotTo(ContainSubstring("javascript:"))
		Expect(output).NotTo(ContainSubstring("data:text/html"))
	})
})
