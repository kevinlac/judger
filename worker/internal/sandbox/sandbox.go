package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type RunOpts struct {
    Image      string
    Mounts     []Mount
    MemoryMB   int
    CPUs       float64
    Network    string
    Args       []string // the command to run inside the container
}

type Mount struct {
    HostPath   string
    Target     string
    ReadOnly   bool
}

// spawns a container, runs it to completion, and returns its output
func RunOnce(ctx context.Context, opts RunOpts) (stdout, stderr string, err error) {
    args := []string{"run", "--rm", "--network", opts.Network}
    if opts.MemoryMB > 0 {
        args = append(args, "--memory", fmt.Sprintf("%dm", opts.MemoryMB))
    }
    if opts.CPUs > 0 {
        args = append(args, "--cpus", fmt.Sprintf("%.1f", opts.CPUs))
    }
    for _, m := range opts.Mounts {
        mount := fmt.Sprintf("type=bind,source=%s,target=%s", m.HostPath, m.Target)
        if m.ReadOnly {
            mount += ",readonly"
        }
        args = append(args, "--mount", mount)
    }
    args = append(args, opts.Image)
    args = append(args, opts.Args...)

    cmd := exec.CommandContext(ctx, "docker", args...)
    var outBuf, errBuf bytes.Buffer
    cmd.Stdout, cmd.Stderr = &outBuf, &errBuf
    err = cmd.Run()
    return outBuf.String(), errBuf.String(), err
}