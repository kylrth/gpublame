package gpublame

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CUDAProcessInfo contains the information provided by `nvidia-smi pmon` for one process.
type CUDAProcessInfo struct {
	Command string
	Type    string
	GPU     int
	PID     int
	Sm      int
	Mem     int
	Enc     int
	Dec     int
}

// There are two process types, (C)ompute and (G)raphics.
const (
	GraphicsProcessType = "G"
	ComputeProcessType  = "C"
)

// Pmon returns a structured version of the output of `nvidia-smi pmon -c 1`. An important
// distinction is that while the command returns prints lines for idle GPUs, Pmon does not return
// anything for them.
func Pmon(ctx context.Context) ([]CUDAProcessInfo, error) {
	var stdout bytes.Buffer

	ps := exec.CommandContext(ctx, "nvidia-smi", "pmon", "-c", "1") // only print once
	ps.Stdout = &stdout

	err := ps.Run()
	if err != nil {
		return nil, fmt.Errorf("execute pmon: %w", err)
	}

	scanner := bufio.NewScanner(&stdout)

	var out []CUDAProcessInfo

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())

		// Skip header lines.
		if strings.HasPrefix(text, "#") {
			continue
		}

		pi, err := parseLine(text)
		if err != nil {
			return out, fmt.Errorf("parse line %d: %w", len(out)+3, err)
		}

		if pi.PID == -1 {
			// idle GPU
			continue
		}

		out = append(out, pi)
	}

	if scanner.Err() == nil {
		return out, ctx.Err()
	}

	return out, scanner.Err()
}

// ParseError can be returned when there was a problem parsing output from a command.
type ParseError struct {
	err  string
	line string
}

// Error returns the error.
func (e *ParseError) Error() string {
	return fmt.Sprintf("%s (line: '%s')", e.err, e.line)
}

func parseLine(line string) (out CUDAProcessInfo, err error) {
	fields := strings.Fields(strings.TrimSpace(line))

	if len(fields) < 8 {
		return CUDAProcessInfo{}, &ParseError{"wrong number of fields in pmon output", line}
	}

	out.GPU, err = strconv.Atoi(fields[0])
	if err != nil {
		return out, fmt.Errorf("parse GPU ID: %w", err)
	}

	out.PID, err = atoi(fields[1], -1)
	if err != nil {
		return out, fmt.Errorf("parse PID: %w", err)
	}

	out.Type = fields[2]

	out.Sm, err = atoi(fields[3], 0)
	if err != nil {
		return out, fmt.Errorf("parse sm: %w", err)
	}

	out.Mem, err = atoi(fields[4], 0)
	if err != nil {
		return out, fmt.Errorf("parse mem: %w", err)
	}

	out.Enc, err = atoi(fields[5], 0)
	if err != nil {
		return out, fmt.Errorf("parse enc: %w", err)
	}

	out.Dec, err = atoi(fields[6], 0)
	if err != nil {
		return out, fmt.Errorf("parse dec: %w", err)
	}

	out.Command = strings.Join(fields[7:], " ")

	return out, nil
}

// atoi is like strconv.Atoi but also "-" maps to `or`.
func atoi(s string, or int) (int, error) {
	if s == "-" {
		return or, nil
	}

	return strconv.Atoi(s)
}
