package loading

import (
	"fmt"
	"io"
	"time"
)

// Indicator prints dots to the given writer at regular intervals
// until Stop is called. It is safe for concurrent use.
type Indicator struct {
	done chan struct{}
	w    io.Writer
}

// Start begins printing dots to w every 500ms.
// Call Stop to end the indicator and print a trailing newline.
func Start(w io.Writer) *Indicator {
	if w == nil {
		w = io.Discard
	}

	ind := &Indicator{
		done: make(chan struct{}),
		w:    w,
	}

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ind.done:
				return
			case <-ticker.C:
				_, _ = fmt.Fprint(ind.w, ".")
			}
		}
	}()

	return ind
}

// Stop ends the loading indicator and prints a trailing newline.
func (ind *Indicator) Stop() {
	close(ind.done)
	_, _ = fmt.Fprintln(ind.w)
}
