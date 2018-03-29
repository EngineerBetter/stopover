package executehelpers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/concourse/fly/ui"
	"github.com/concourse/go-archive/tgzfs"
	"github.com/concourse/go-concourse/concourse"
)

func Upload(client concourse.Client, input Input, includeIgnored bool) {
	path := input.Path
	pipe := input.Pipe

	var files []string
	var err error

	if includeIgnored {
		files = []string{"."}
	} else {
		files, err = getGitFiles(path)
		if err != nil {
			files = []string{"."}
		}
	}

	archiveStream, archiveWriter := io.Pipe()

	go func() {
		archiveWriter.CloseWithError(tgzfs.Compress(archiveWriter, path, files...))
	}()

	upload, err := http.NewRequest("PUT", pipe.WriteURL, archiveStream)
	if err != nil {
		panic(err)
	}

	response, err := client.HTTPClient().Do(upload)
	if err != nil {
		fmt.Fprintln(ui.Stderr, "upload request failed:", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Fprintln(ui.Stderr, badResponseError("uploading bits", response))
	}
}

func getGitFiles(dir string) ([]string, error) {
	tracked, err := gitLS(dir)
	if err != nil {
		return nil, err
	}

	untracked, err := gitLS(dir, "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	return append(tracked, untracked...), nil
}

func gitLS(dir string, flags ...string) ([]string, error) {
	files := []string{}

	gitLS := exec.Command("git", append([]string{"ls-files", "-z"}, flags...)...)
	gitLS.Dir = dir

	gitOut, err := gitLS.StdoutPipe()
	if err != nil {
		return nil, err
	}

	outScan := bufio.NewScanner(gitOut)
	outScan.Split(scanNull)

	err = gitLS.Start()
	if err != nil {
		return nil, err
	}

	for outScan.Scan() {
		files = append(files, outScan.Text())
	}

	err = gitLS.Wait()
	if err != nil {
		return nil, err
	}

	return files, nil
}

func scanNull(data []byte, atEOF bool) (int, []byte, error) {
	// eof, no more data; terminate
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// look for terminating null byte
	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// no final terminator; return what's left
	if atEOF {
		return len(data), data, nil
	}

	// request more data
	return 0, nil, nil
}
