package buildah

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
)

type DockerWithFuseBuildah struct {
	BaseBuildah

	HostStorageDir string
}

func NewDockerWithFuseBuildah() (*DockerWithFuseBuildah, error) {
	b := &DockerWithFuseBuildah{}

	baseBuildah, err := NewBaseBuildah()
	if err != nil {
		return nil, fmt.Errorf("unable to create BaseBuildah: %s", err)
	}
	b.BaseBuildah = *baseBuildah

	b.HostStorageDir = filepath.Join(werf.GetHomeDir(), "buildah", "storage")
	if err := os.MkdirAll(b.HostStorageDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", b.HostStorageDir, err)
	}

	return b, nil
}

func (b *DockerWithFuseBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	sessionTmpDir, _, _, err := b.prepareBuildFromDockerfile(dockerfile, opts.ContextTar)
	if err != nil {
		return "", fmt.Errorf("error preparing for build from dockerfile: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(sessionTmpDir); err != nil {
			logboek.Warn().LogF("unable to remove session tmp dir %q: %s\n", sessionTmpDir, err)
		}
	}()

	output, _, err := b.runBuildah(
		ctx,
		[]string{
			"--volume", fmt.Sprintf("%s:/.werf/buildah/tmp", sessionTmpDir),
			"--workdir", "/.werf/buildah/tmp/context",
		},
		[]string{"bud", "-f", "/.werf/buildah/tmp/Dockerfile"}, opts.LogWriter,
	)
	if err != nil {
		return "", err
	}

	outputLines := scanLines(output)

	return outputLines[len(outputLines)-1], nil
}

func (b *DockerWithFuseBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, append([]string{"run", container}, command...), opts.LogWriter)
	return err
}

// FIXME(ilya-lesikov): add to interface
func (b *DockerWithFuseBuildah) FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, []string{"from", "--name", container, image}, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) runBuildah(ctx context.Context, dockerArgs []string, buildahArgs []string, logWriter io.Writer) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	var stdoutWriter io.Writer
	var stderrWriter io.Writer

	if logWriter != nil {
		stdoutWriter = io.MultiWriter(stdout, logWriter)
		stderrWriter = io.MultiWriter(stderr, logWriter)
	} else {
		stdoutWriter = stdout
		stderrWriter = stderr
	}

	args := []string{"--rm"}
	args = append(args, dockerArgs...)
	args = append(args, buildahWithFuseDockerArgs(b.HostStorageDir)...)
	args = append(args, buildahArgs...)

	fmt.Printf("ARGS: %v\n", args)

	err := docker.CliRun_ProvidedOutput(ctx, stdoutWriter, stderrWriter, args...)
	return stdout.String(), stderr.String(), err
}

func buildahWithFuseDockerArgs(hostStorageDir string) []string {
	return []string{
		"--device", "/dev/fuse",
		"--security-opt", "seccomp=unconfined",
		"--security-opt", "apparmor=unconfined",
		"--volume", fmt.Sprintf("%s:/var/lib/containers/storage", hostStorageDir),
		BuildahImage, "buildah",
	}
}

func scanLines(data string) []string {
	var lines []string

	s := bufio.NewScanner(strings.NewReader(data))
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines
}