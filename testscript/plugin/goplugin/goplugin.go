// Package goplugin implements the reference "go" testscript plugin. It starts
// a [goproxytest] module proxy serving the modules found in the plugin
// parameter directory and provides a "go" command that runs the real go tool
// against that proxy in a hermetic environment.
//
// The cmd/testscript-plugin-go binary is a thin wrapper around [Run].
package goplugin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/rogpeppe/go-internal/goproxytest"
	"github.com/rogpeppe/go-internal/testscript/plugin"
)

// Run runs the go plugin server. It is the whole body of the
// cmd/testscript-plugin-go binary.
func Run() error {
	return plugin.Serve(&goPlugin{})
}

// envVars holds the names of the environment variables the go command needs
// passed to it. It is a superset of resultEnvVars: the extra variables (PATH,
// HOME, TMPDIR, ...) are provided by the test environment rather than the
// plugin, but are still needed for the go tool to run.
var envVars = append([]string{
	"PATH", "HOME", "TMPDIR", "TEMP", "TMP", "SystemRoot", "windir",
}, resultEnvVars...)

// resultEnvVars holds the names of the environment variables that the plugin
// sets in the test environment.
var resultEnvVars = []string{
	"GOROOT", "GOCACHE", "GOPATH", "GOARCH", "GOOS",
	"CCACHE_DISABLE", "GOPROXY", "GONOSUMDB", "GOSUMDB",
}

type goPlugin struct{}

func (*goPlugin) Info() plugin.PluginInfo {
	requiredEnv := map[string]bool{"WORK": true}
	cmdEnv := make(map[string]bool)
	for _, v := range envVars {
		cmdEnv[v] = true
	}
	resultEnv := make(map[string]bool)
	for _, v := range resultEnvVars {
		resultEnv[v] = true
	}
	return plugin.PluginInfo{
		RequiredEnv: requiredEnv,
		ResultEnv:   resultEnv,
		Cmds: map[string]plugin.CmdInfo{
			"go": {
				RequiredEnv: cmdEnv,
				NeedStdout:  true,
				NeedStderr:  true,
			},
		},
	}
}

func (*goPlugin) NewTestInstance(p plugin.TestParams) (plugin.TestInstance, error) {
	if info, err := os.Stat(p.PluginParamsDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("no go proxy directory found at %q", p.PluginParamsDir)
	}
	work := p.Environ["WORK"]
	if work == "" {
		return nil, fmt.Errorf("the go plugin requires the WORK environment variable")
	}
	base, err := baseEnv()
	if err != nil {
		return nil, err
	}
	srv, err := goproxytest.NewServer(p.PluginParamsDir, "")
	if err != nil {
		return nil, fmt.Errorf("cannot start go proxy: %v", err)
	}
	return &goPluginInstance{
		srv:  srv,
		work: work,
		base: base,
	}, nil
}

func (*goPlugin) Close() {}

type goPluginInstance struct {
	srv  *goproxytest.Server
	work string
	base goBaseEnv
}

func (inst *goPluginInstance) Env() map[string]string {
	return map[string]string{
		"GOROOT":         inst.base.GOROOT,
		"GOCACHE":        inst.base.GOCACHE,
		"GOPATH":         filepath.Join(inst.work, ".gopath"),
		"GOARCH":         runtime.GOARCH,
		"GOOS":           runtime.GOOS,
		"CCACHE_DISABLE": "1",
		"GOPROXY":        inst.srv.URL,
		"GONOSUMDB":      "*",
		"GOSUMDB":        "off",
	}
}

func (inst *goPluginInstance) RunCmd(p plugin.CmdParams) (plugin.CmdResult, error) {
	if p.CmdName != "go" {
		return plugin.CmdResult{}, fmt.Errorf("unrecognized command %q", p.CmdName)
	}
	cmd := exec.Command("go", p.Args...)
	cmd.Dir = p.CWD
	cmd.Env = make([]string, 0, len(p.Environ))
	for k, v := range p.Environ {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdin = bytes.NewReader(p.Stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	res := plugin.CmdResult{}
	if err := cmd.Run(); err != nil {
		// A failure of the go command itself is reported via Err so the
		// host can honour command negation.
		res.Err = err.Error()
	}
	res.Stdout = stdout.Bytes()
	res.Stderr = stderr.Bytes()
	return res, nil
}

func (inst *goPluginInstance) Close() {
	inst.srv.Close()
}

// goBaseEnv holds the parts of the go environment that are constant across
// test instances.
type goBaseEnv struct {
	GOROOT  string
	GOCACHE string
}

var (
	baseEnvOnce sync.Once
	baseEnvVal  goBaseEnv
	baseEnvErr  error
)

// baseEnv returns the constant parts of the go environment, computed once.
func baseEnv() (goBaseEnv, error) {
	baseEnvOnce.Do(func() {
		out, err := exec.Command("go", "env", "GOROOT", "GOCACHE").Output()
		if err != nil {
			baseEnvErr = fmt.Errorf("cannot run go env: %v", err)
			return
		}
		fields := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
		if len(fields) != 2 {
			baseEnvErr = fmt.Errorf("unexpected go env output %q", out)
			return
		}
		baseEnvVal = goBaseEnv{GOROOT: fields[0], GOCACHE: fields[1]}
	})
	return baseEnvVal, baseEnvErr
}
