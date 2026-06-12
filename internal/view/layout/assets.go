package layout

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
)

// cssPath is the on-disk location of the generated stylesheet, relative to the
// server's working directory (the repo root, matching e.Static("/static", "static")).
const cssPath = "static/css/output.css"

var (
	cssVerOnce sync.Once
	cssVer     string
)

// cssVersion returns a short content hash of the generated CSS, used as a
// cache-busting query string so browsers refetch the stylesheet whenever it
// changes instead of serving a stale cached copy.
func cssVersion() string {
	cssVerOnce.Do(func() {
		data, err := os.ReadFile(cssPath)
		if err != nil {
			cssVer = "dev"
			return
		}
		sum := sha256.Sum256(data)
		cssVer = hex.EncodeToString(sum[:])[:8]
	})
	return cssVer
}
