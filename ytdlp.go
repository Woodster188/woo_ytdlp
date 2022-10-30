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
	OutFilename  string
	VideoExt     string
	AudioExt     string
	VideoQuality int //480, 720, 1080 etc
}

type ytdlp struct {
	Path    string
	Options YtdlpOptions
}

type Ytdlp interface {
	Download(ctx context.Context, link, to string, prgressCh chan int) error
	SetOutFilename(filename string)
	SetQuality(quality int)
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
	cmd.Dir = toDir
	
	stdoutPipe, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanWords)
	go func() {
		var progress int
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "%") {
				newProgress := formatProgress(line) / 2
				if progress != newProgress {
					progress = newProgress
					progressCh <- progress
				}
			}
		}
	}()
	err := startWithErr(cmd)
	err = WaitWithErr(cmd)
	return err
}

func (yt *ytdlp) SetOutFilename(filename string) {
	yt.Options.OutFilename = filename
}
func (yt *ytdlp) SetQuality(quality int) {
	yt.Options.VideoQuality = quality
}

func formatProgress(prStr string) int {
	prStr = strings.Replace(prStr, "%", "", -1)
	var result float64
	result, _ = strconv.ParseFloat(prStr, 64)
	return int(result)

}

func NewYtdlp(path string) Ytdlp {
	options := YtdlpOptions{
		OutFilename:  "video.mp4",
		VideoExt:     "mp4",
		AudioExt:     "m4a",
		VideoQuality: 1080,
	}
	return &ytdlp{Path: path, Options: options}
}

func startWithErr(cmd *exec.Cmd) error {
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	err := cmd.Start()
	if err != nil {
		return errors.New(errBuf.String())
	}
	return nil
}

func waitWithErr(cmd *exec.Cmd) error {
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	err := cmd.Wait()
	if err != nil {
		return errors.New(errBuf.String())
	}
	return nil
}
