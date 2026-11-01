package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

var renderer = goldmark.New(goldmark.WithRendererOptions(html.WithHardWraps()))

func Render(source string) (string, error) {
	var output bytes.Buffer
	if err := renderer.Convert([]byte(source), &output); err != nil {
		return "", err
	}
	return output.String(), nil
}
