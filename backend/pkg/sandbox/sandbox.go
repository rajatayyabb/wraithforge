package sandbox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Limits struct {
	MemoryBytes int64
	CPUShares   int64
	Timeout     time.Duration
}

func DefaultLimits() Limits {
	return Limits{
		MemoryBytes: 256 * 1024 * 1024,
		CPUShares:   512,
		Timeout:     60 * time.Second,
	}
}

type Runner struct {
	cli   *client.Client
	image string
}

func NewRunner(image string) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Runner{cli: cli, image: image}, nil
}

func (r *Runner) Close() error {
	return r.cli.Close()
}

func (r *Runner) Run(ctx context.Context, cmd []string, limits Limits, out chan<- string) error {
	defer close(out)

	ctx, cancel := context.WithTimeout(ctx, limits.Timeout)
	defer cancel()

	resp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image:      r.image,
		Cmd:        cmd,
		Tty:        false,
		StopSignal: "SIGKILL",
	}, &container.HostConfig{
		AutoRemove: false,
		Resources: container.Resources{
			Memory:    limits.MemoryBytes,
			CPUShares: limits.CPUShares,
		},
	}, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	containerID := resp.ID

	defer func() {
		removeCtx, removeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer removeCancel()
		_ = r.cli.ContainerRemove(removeCtx, containerID, container.RemoveOptions{Force: true})
	}()

	if err := r.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	logsReader, err := r.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("attach logs: %w", err)
	}
	defer logsReader.Close()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, _ = stdcopy.StdCopy(pw, pw, logsReader)
	}()

	scanner := bufio.NewScanner(pr)
	go func() {
		for scanner.Scan() {
			select {
			case out <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	statusCh, errCh := r.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			killCtx, killCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer killCancel()
			_ = r.cli.ContainerKill(killCtx, containerID, "SIGKILL")
			return err
		}
	case <-statusCh:
	case <-ctx.Done():
		killCtx, killCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer killCancel()
		_ = r.cli.ContainerKill(killCtx, containerID, "SIGKILL")
		return ctx.Err()
	}

	return nil
}
