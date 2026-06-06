package logic

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"
)

func GC(w io.Writer) error {
	w.Write([]byte(time.Now().Format("2006/01/02 15:04:05") + "\n"))
	printMemStats(w, "GC Before")
	runtime.GC()
	printMemStats(w, "GC After")
	return nil
}

func printMemStats(w io.Writer, header string) {
	fmt.Fprintln(w, header)
	buf := getMemStats()
	fmt.Fprintln(w, buf)
}

func getMemStats() string {

	var buf strings.Builder
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	buf.WriteString(fmt.Sprintf("----------------------------------\n"))
	buf.WriteString(fmt.Sprintf("| Alloc      :%.1fMiB\n", toMiB(ms.Alloc)))
	buf.WriteString(fmt.Sprintf("| HeapAlloc  :%.1fMiB\n", toMiB(ms.HeapAlloc)))
	buf.WriteString(fmt.Sprintf("| Sys        :%.1fMiB\n", toMiB(ms.Sys)))
	buf.WriteString(fmt.Sprintf("----------------------------------"))

	return buf.String()
}

func toKiB(v uint64) float64 {
	return float64(v) / 1024
}

func toMiB(v uint64) float64 {
	return toKiB(v) / 1024
}
