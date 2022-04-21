package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kylrth/gpublame"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}

var cmd = &cobra.Command{
	Short: "Determine who's using the GPUs",
	Args:  cobra.NoArgs,
	Run: func(*cobra.Command, []string) {
		err := blame(verbosity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	},
}

var verbosity int

func init() {
	cmd.Flags().IntVarP(&verbosity, "verbosity", "v", 1, "Set desired verbosity (0-3). Verbosity "+
		">= 2 includes the PGID of the process, and >= 3 includes the command that started the "+
		"process group.")
}

// blame prints out who's using each GPU.
func blame(verbosity int) error {
	processes, err := gpublame.Pmon(context.Background())
	if err != nil {
		return fmt.Errorf("get GPU processes: %w", err)
	}

	// a map from GPU ID to process groups using it.
	usage := make(map[int][]gpublame.ProcessInfo)

	for _, process := range processes {
		if process.Type != gpublame.ComputeProcessType {
			continue
		}

		var pi gpublame.ProcessInfo

		pi, err = gpublame.ProcessGroupInfo(process.PID)
		if err != nil {
			return fmt.Errorf("get process group info for PID %d: %w", process.PID, err)
		}

		usage[process.GPU] = append(usage[process.GPU], pi)
	}

	numGPUs, err := gpublame.CountGPUs(context.Background())
	if err != nil {
		return fmt.Errorf("get number of GPUs: %w", err)
	}

	printResults(usage, numGPUs, verbosity)

	return nil
}

func printResults(usage map[int][]gpublame.ProcessInfo, numGPUs, verbosity int) {
	maxID := numGPUs - 1
	now := time.Now()

	fmt.Println("GPU Users:")

	for id := 0; id <= maxID; id++ {
		if len(usage[id]) == 0 {
			fmt.Printf("%d:\n", id)

			continue
		}

		var b strings.Builder

		b.WriteString(fmt.Sprintf("%d:", id))

		for _, pi := range usage[id] {
			b.WriteRune(' ')
			b.WriteString(pi.User)
			b.WriteRune('(')
			b.WriteString(durationPrint(now.Sub(pi.Start)))

			if verbosity > 1 {
				b.WriteString(fmt.Sprintf(",pgid=%d", pi.PID))

				if verbosity > 2 {
					b.WriteString(`,cmd="` + pi.Command + `"`)
				}
			}

			b.WriteString("),")
		}

		fmt.Println(b.String()[:b.Len()-1])
	}
}

// durationPrint is like (time.Duration).String() with the addition of days (e.g. "1d6h55m10s").
func durationPrint(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days == 0 {
		return d.Round(time.Second).String()
	}

	d -= time.Hour * time.Duration(24*days)

	return fmt.Sprintf("%dd%s", days, d.Round(time.Second).String())
}
