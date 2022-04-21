package gpublame

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PGID returns the process group ID for a specified process.
func PGID(pid int) (int, error) {
	var stdout bytes.Buffer

	//nolint:gosec // The PID will always be an int, so this is safe.
	ps := exec.Command("ps", "-heo", "pgid", "-q", strconv.Itoa(pid))
	ps.Stdout = &stdout

	if err := ps.Run(); err != nil {
		return 0, fmt.Errorf("ps: %w", err)
	}

	return strconv.Atoi(strings.TrimSpace(stdout.String()))
}

// ProcessInfo is the information returned by ps for a single process.
type ProcessInfo struct {
	Start   time.Time
	User    string
	Command string
	PID     int
}

// ProcessGroupInfo returns information provided by ps for the process group of the specified
// process.
func ProcessGroupInfo(pid int) (ProcessInfo, error) {
	pgid, err := PGID(pid)
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("get PGID: %w", err)
	}

	var stdout bytes.Buffer

	//nolint:gosec // The PGID will always be an int, so this is safe.
	ps := exec.Command("ps", "-heo", "user,lstart,cmd", "-q", strconv.Itoa(pgid))
	ps.Stdout = &stdout

	err = ps.Run()
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("find user for PID %d: %w", pid, err)
	}

	return processInfoFromFields(stdout.String(), pgid)
}

func processInfoFromFields(s string, pid int) (p ProcessInfo, err error) {
	fields := strings.Fields(s)

	if len(fields) < 7 {
		return p, &ParseError{fmt.Sprintf("expected 7 fields; got %d", len(fields)), s}
	}

	p.Start, err = time.Parse("Mon Jan _2 15:04:05 2006", strings.Join(fields[1:6], " "))
	if err != nil {
		return p, fmt.Errorf("invalid time field: %w", err)
	}

	p.User = fields[0]
	p.PID = pid
	p.Command = strings.Join(fields[6:], " ")

	return p, nil
}
