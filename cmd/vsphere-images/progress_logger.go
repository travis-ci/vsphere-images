package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/progress"
)

type progressLogger struct {
	prefix string
	wg     sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func newProgressLogger(prefix string) *progressLogger {
	p := &progressLogger{
		prefix: prefix,

		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(tick, ch)
			if err != nil {
				stop = true
			}
		case <-p.done:
			stop = true
		case <-tick.C:
			fmt.Fprintf(os.Stderr, "\r%s      ", p.prefix)
		}
	}

	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "\r%s      ", p.prefix)
		fmt.Fprintf(os.Stderr, "\r%sError: %s\n", p.prefix, err)
	} else {
		fmt.Fprintf(os.Stderr, "\r%s      ", p.prefix)
		fmt.Fprintf(os.Stderr, "\r%sOK\n", p.prefix)
	}
}

func (p *progressLogger) loopB(tick *time.Ticker, ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		case <-tick.C:
			line := "\r" + p.prefix
			if r != nil {
				line += fmt.Sprintf("(%.0f%%", r.Percentage())
				detail := r.Detail()
				if detail != "" {
					line += fmt.Sprintf(", %s", detail)
				}
				line += ")"
			}
			fmt.Fprintf(os.Stderr, "%s", line)
		}
	}

	return err
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
}

func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}
