package static

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed styles/*.css
var styles embed.FS

func Styles() (fs.FS, error) {
	styles, err := fs.Sub(styles, "styles")
	if err != nil {
		return nil, fmt.Errorf("fs sub: %v", err)
	}

	return styles, nil
}
