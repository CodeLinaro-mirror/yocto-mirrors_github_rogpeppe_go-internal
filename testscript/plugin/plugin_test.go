package plugin_test

import (
	"os/exec"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/rogpeppe/go-internal/testscript/plugin"
	"github.com/rogpeppe/go-internal/testscript/plugin/goplugin"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"testscript-plugin-go": func() {
			if err := goplugin.Run(); err != nil {
				panic(err)
			}
		},
	})
}

func TestGoPlugin(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}
	p := testscript.Params{
		Dir: "testdata",
	}
	cleanup, err := plugin.Setup(&p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	testscript.Run(t, p)
}
