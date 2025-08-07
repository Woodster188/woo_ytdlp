package woo_ytdlp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type YtdlpOptions struct {
	OutFilename   string
	VideoExt      string
	AudioExt      string
	VideoQuality  int //480, 720, 1080 etc
	ErrWithOutput bool
	CookiePath    string
}

type ytdlp struct {
	Path    string
	Options YtdlpOptions
}

type Ytdlp interface {
	Download(ctx context.Context, link, to string, prgressCh chan int) error
	SetOutFilename(filename string)
	SetQuality(quality int)
	SetErrWithOutput(mode bool)
	SetCookiePath(path string)
}

func (yt *ytdlp) Download(ctx context.Context, link, toDir string, progressCh chan int) error {
	if yt.Path == "" {
		return errors.New("path to yt-dlp is not configured")
	}
	cmd := exec.CommandContext(
		ctx,
		yt.Path,
	)
	ytFiltersStr := fmt.Sprintf(
		"bestvideo[ext=%s][height=%d]+bestaudio[ext=%s]/bestvideo[ext=%s][height=%d]+bestaudio/best",
		yt.Options.VideoExt, yt.Options.VideoQuality, yt.Options.AudioExt, yt.Options.VideoExt, yt.Options.VideoQuality,
	)
	cmd.Args = append(cmd.Args, link)
	cmd.Args = append(cmd.Args, "-f", ytFiltersStr)
	cmd.Args = append(cmd.Args, "-o", yt.Options.OutFilename)
	if yt.Options.CookiePath != "" {
		cmd.Args = append(cmd.Args, "--cookies", yt.Options.CookiePath)
	}
	cmd.Dir = toDir

	stdoutPipe, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanWords)
	var outBuf strings.Builder
	go func() {
		var progress int
		for scanner.Scan() {
			line := scanner.Text()
			if yt.Options.ErrWithOutput {
				outBuf.WriteString(line)
			}
			if strings.Contains(line, "%") {
				newProgress := formatProgress(line) / 2
				if progress != newProgress {
					progress = newProgress
					progressCh <- progress
				}
			}
		}
	}()
	err := cmd.Start()
	err = cmd.Wait()
	if err != nil && yt.Options.ErrWithOutput {
		err = fmt.Errorf("%s: %s", err.Error(), outBuf.String())
	}
	return err
}

func (yt *ytdlp) SetOutFilename(filename string) {
	yt.Options.OutFilename = filename
}
func (yt *ytdlp) SetQuality(quality int) {
	yt.Options.VideoQuality = quality
}

func (yt *ytdlp) SetErrWithOutput(mode bool) {
	yt.Options.ErrWithOutput = mode
}

func (yt *ytdlp) SetCookiePath(path string) {
	yt.Options.CookiePath = path
}

func formatProgress(prStr string) int {
	prStr = strings.Replace(prStr, "%", "", -1)
	var result float64
	result, _ = strconv.ParseFloat(prStr, 64)
	return int(result)

}

func NewYtdlp(path string) Ytdlp {
	options := YtdlpOptions{
		OutFilename:   "video.mp4",
		VideoExt:      "mp4",
		AudioExt:      "m4a",
		VideoQuality:  1080,
		ErrWithOutput: false,
		CookiePath:    "",
	}
	return &ytdlp{Path: path, Options: options}
}
