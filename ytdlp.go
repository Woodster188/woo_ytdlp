package ytdlp_wrapper

import (
	"bufio"
	"context"
	"errors"
	"os/exec"
	"strings"
)

type ytdlp struct {
	Path string
}

type Ytdlp interface {
	Download(ctx context.Context, link, to string, prgressCh chan string) error
}

func (yt *ytdlp) Download(ctx context.Context, link, toDir string, progressCh chan string) error {
	defer close(progressCh)
	if yt.Path == "" {
		return errors.New("path to yt-dlp is not configured")
	}
	cmd := exec.CommandContext(
		ctx,
		yt.Path,
	)
	cmd.Args = append(cmd.Args, link)
	cmd.Args = append(cmd.Args, "-f", "bestvideo[ext=mp4][height=1080]+bestaudio[ext=m4a]/bestvideo+bestaudio/best")
	cmd.Args = append(cmd.Args, "-o", "video.mp4")
	cmd.Dir = toDir

	stdoutPipe, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanWords)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "%") {
				progressCh <- line
			}
		}
	}()
	err := cmd.Start()
	err = cmd.Wait()
	return err
}

func NewYtdlp(path string) Ytdlp {
	return &ytdlp{Path: path}
}
