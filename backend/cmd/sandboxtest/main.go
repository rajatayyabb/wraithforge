package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"wraithforge/backend/pkg/sandbox"
)

func main() {
	image := flag.String("image", "alpine:3.20", "base image")
	command := flag.String("cmd", "echo hello", "shell command to run")
	timeout := flag.Duration("timeout", 60*time.Second, "hard timeout")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("signal received, cancelling")
		cancel()
	}()

	runner, err := sandbox.NewRunner(*image)
	if err != nil {
		fmt.Println("runner init error:", err)
		os.Exit(1)
	}
	defer runner.Close()

	limits := sandbox.DefaultLimits()
	limits.Timeout = *timeout

	out := make(chan string)
	go func() {
		for line := range out {
			fmt.Println(line)
		}
	}()

	cmd := []string{"sh", "-c", strings.TrimSpace(*command)}
	if err := runner.Run(ctx, cmd, limits, out); err != nil {
		fmt.Println("run error:", err)
		os.Exit(1)
	}
}
