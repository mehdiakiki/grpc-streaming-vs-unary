package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func peakAllocWhile(fn func()) uint64 {
	var peak uint64
	done := make(chan struct{})

	go func() {
		t := time.NewTicker(10 * time.Millisecond)
		defer t.Stop()
		var ms runtime.MemStats
		for {
			select {
			case <-done:
				return
			case <-t.C:
				runtime.ReadMemStats(&ms)
				if ms.Alloc > peak {
					peak = ms.Alloc
				}
			}
		}
	}()

	fn()
	close(done)
	return peak
}

func run(mode, file string) (time.Duration, uint64) {
	var dur time.Duration
	peak := peakAllocWhile(func() {
		start := time.Now()
		cmd := exec.Command("go", "run", "./cmd/client", "-mode", mode, "-file", file, "-chunk", "1048576")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		dur = time.Since(start)
	})
	return dur, peak
}

func main() {
	file := "test.bin"

	d1, p1 := run("unary", file)
	d2, p2 := run("stream", file)

	log.Printf("RESULT unary  : elapsed=%s peak_alloc=%d MB", d1, p1/1024/1024)
	log.Printf("RESULT stream : elapsed=%s peak_alloc=%d MB", d2, p2/1024/1024)
}
