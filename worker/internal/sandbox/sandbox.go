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

func buildRunFlags(opts RunOpts) []string {
    args := []string{"run", "--rm", "--network", opts.Network}
    if opts.MemoryMB > 0 {
        args = append(args, "--memory", fmt.Sprintf("%dm", opts.MemoryMB))
    }
    if opts.CPUs > 0 {
        args = append(args, "--cpus", fmt.Sprintf("%.2f", opts.CPUs))
    }
    for _, m := range opts.Mounts {
        mount := fmt.Sprintf("type=bind,source=%s,target=%s", m.HostPath, m.Target)
        if m.ReadOnly {
            mount += ",readonly"
        }
        args = append(args, "--mount", mount)
    }
    return args // flags only, no image/command yet
}

// spawns a container, runs it to completion, and returns its output
func RunOnce(ctx context.Context, opts RunOpts) (stdout, stderr string, err error) {
    args := buildRunFlags(opts)
    args = append(args, opts.Image)
    args = append(args, opts.Args...)

    cmd := exec.CommandContext(ctx, "docker", args...)
    var outBuf, errBuf bytes.Buffer
    cmd.Stdout, cmd.Stderr = &outBuf, &errBuf
    err = cmd.Run()
    return outBuf.String(), errBuf.String(), err
}

func Kill(containerName string) error {
    return exec.Command("docker", "rm", "-f", containerName).Run()
}

func Exec(ctx context.Context, containerName string, args []string, stdin []byte) (stdout, stderr string, err error) {
    fullArgs := append([]string{"exec", "-i", containerName}, args...)
    cmd := exec.CommandContext(ctx, "docker", fullArgs...)
    cmd.Stdin = bytes.NewReader(stdin)

    var outBuf, errBuf bytes.Buffer
    cmd.Stdout, cmd.Stderr = &outBuf, &errBuf
    err = cmd.Run()
    return outBuf.String(), errBuf.String(), err
}

func Spawn(ctx context.Context, containerName string, opts RunOpts) error {
    args := append(buildRunFlags(opts), "-d", "--name", containerName)
    args = append(args, opts.Image, "tail", "-f", "/dev/null")

    cmd := exec.CommandContext(ctx, "docker", args...)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("spawn: %w: %s", err, stderr.String())
    }
    return nil
}