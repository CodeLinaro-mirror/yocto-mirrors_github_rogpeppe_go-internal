// Command testscript-plugin-go is the reference "go" plugin for the testscript
// command. It starts a module proxy serving the modules in the directory
// passed to the "plugin go <dir>" command and provides a "go" command that
// runs the go tool against that proxy.
package main

import (
	"fmt"
	"os"

	"github.com/rogpeppe/go-internal/testscript/plugin/goplugin"
)

func main() {
	if err := goplugin.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "testscript-plugin-go:", err)
		os.Exit(1)
	}
}
