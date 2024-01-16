package gpublame

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// CountGPUs returns the number of GPUs according to `nvidia-smi`.
func CountGPUs(ctx context.Context) (int, error) {
	var stdout bytes.Buffer

	smi := exec.CommandContext(ctx, "nvidia-smi", "-L")
	smi.Stdout = &stdout
	
	err := smi.Run()
	if err != nil {
		return 0, fmt.Errorf("run nvidia-smi: %w", err)
	}

	cnt, err := lineCounter(&stdout)
	if err != nil {
		return cnt, fmt.Errorf("count lines lol: %w", err)
	}

	return cnt, nil
}

//nolint:lll // URL
// lineCounter is taken from
// https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently.
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
