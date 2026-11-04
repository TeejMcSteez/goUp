// Command rtwatch fetches a /api/rt snapshot from a goUp server, then polls
// append frames and prints the merged series. It follows the same fetch and
// merge steps as frontend/src/hooks/useResponseTimeSeries.ts.
//
//	go run ./cmd/rtwatch                              # localhost, poll every 5s
//	go run ./cmd/rtwatch -url https://goup.example.com
//	go run ./cmd/rtwatch -fire -interval 10s          # run checks before each poll
//	go run ./cmd/rtwatch -once -tail 0                # full snapshot, then exit
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"goUp/server"
	"io"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const usPerMs = 1000

// series is the merged history. Cols holds microseconds, 0 = no sample.
type series struct {
	names []string
	seq   uint32
	ts    []uint32
	cols  [][]uint32
}

type options struct {
	base     string
	interval time.Duration
	tail     int
	once     bool
	fire     bool
}

func main() {
	var opts options
	flag.StringVar(&opts.base, "url", "http://localhost:8101", "goUp base URL")
	flag.DurationVar(&opts.interval, "interval", 5*time.Second, "time between polls")
	flag.IntVar(&opts.tail, "tail", 10, "rows to show per table, 0 = all")
	flag.BoolVar(&opts.once, "once", false, "exit after the snapshot")
	flag.BoolVar(&opts.fire, "fire", false, "POST /api/fire before each poll (runs real service checks)")
	flag.Parse()

	opts.base = strings.TrimRight(opts.base, "/")
	if !strings.Contains(opts.base, "://") {
		opts.base = "http://" + opts.base
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, opts); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "error: %s: %v\n", opts.base, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, opts options) error {
	var cur *series
	for {
		var err error
		if cur, err = poll(ctx, opts, cur); err != nil {
			return err
		}
		if cur != nil {
			fmt.Printf("[merged] seq=%d rows=%d cols=%d\n", cur.seq, len(cur.ts), len(cur.names))
			printTable(cur.names, cur.ts, cur.cols, opts.tail)
		}
		fmt.Println()

		if opts.once {
			return nil
		}
		if err := sleep(ctx, opts.interval); err != nil {
			return err
		}
		if opts.fire {
			if _, err := request(ctx, opts.base+"/api/fire", struct{}{}); err != nil {
				return err
			}
			// The check runs asynchronously; give it a moment to store rows.
			if err := sleep(ctx, time.Second); err != nil {
				return err
			}
		}
	}
}

// poll advances prev by one fetch: a snapshot when there's nothing to extend,
// otherwise an append, falling back to a snapshot on an unknown service.
func poll(ctx context.Context, opts options, prev *series) (*series, error) {
	if prev == nil {
		return fetchSnapshot(ctx, opts)
	}

	frame, err := fetchFrame(ctx, opts, map[string]uint32{"from": prev.ts[len(prev.ts)-1]})
	if err != nil {
		return nil, err
	}
	printFrame(frame, opts.tail)
	if len(frame.Ts) == 0 {
		return prev, nil
	}

	colMap := mapColumns(prev, frame)
	if colMap == nil {
		fmt.Printf("unknown service in %v, falling back to snapshot\n", frame.Names)
		return fetchSnapshot(ctx, opts)
	}
	return mergeFrame(prev, frame, colMap), nil
}

func fetchSnapshot(ctx context.Context, opts options) (*series, error) {
	frame, err := fetchFrame(ctx, opts, nil)
	if err != nil {
		return nil, err
	}
	if frame.Typ != server.FrameSnapshot {
		return nil, errors.New("expected a snapshot frame from /api/rt")
	}
	printFrame(frame, opts.tail)
	if len(frame.Ts) == 0 || len(frame.Names) == 0 {
		return nil, nil
	}
	return &series{frame.Names, frame.Seq, frame.Ts, frame.Cols}, nil
}

// fetchFrame GETs /api/rt, or POSTs body to it when body is non-nil.
func fetchFrame(ctx context.Context, opts options, body map[string]uint32) (server.Frame, error) {
	var payload any
	if body != nil {
		payload = body
	}
	buf, err := request(ctx, opts.base+"/api/rt", payload)
	if err != nil {
		return server.Frame{}, err
	}
	return server.DecodeFrame(buf)
}

// request GETs url, or POSTs payload to it as JSON when payload is non-nil.
func request(ctx context.Context, url string, payload any) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	method := http.MethodGet
	var body io.Reader
	if payload != nil {
		method = http.MethodPost
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s: %s: %s", method, url, res.Status, bytes.TrimSpace(buf))
	}
	return buf, nil
}

// mapColumns maps each frame column to its column index in prev by name, since
// a window only holds the services that reported in it. Returns nil when the
// frame has a service prev doesn't know about.
func mapColumns(prev *series, frame server.Frame) []int {
	colMap := make([]int, len(frame.Names))
	for c, name := range frame.Names {
		colMap[c] = slices.Index(prev.names, name)
		if colMap[c] == -1 {
			return nil
		}
	}
	return colMap
}

// mergeFrame appends the frame's rows onto prev without mutating it. The window
// starts at prev's last timestamp (inclusive), so a row at that same second
// merges into prev's last row instead of duplicating it.
func mergeFrame(prev *series, frame server.Frame, colMap []int) *series {
	prevLen := len(prev.ts)
	lastTs := prev.ts[prevLen-1]

	start := 0
	for start < len(frame.Ts) && frame.Ts[start] <= lastTs {
		start++
	}
	added := len(frame.Ts) - start

	cols := make([][]uint32, len(prev.cols))
	for c, col := range prev.cols {
		cols[c] = append(slices.Clone(col), make([]uint32, added)...)
	}
	for c, col := range frame.Cols {
		target := cols[colMap[c]]
		for i, v := range col {
			if v == 0 || frame.Ts[i] < lastTs {
				continue
			}
			if i < start {
				target[prevLen-1] = v
			} else {
				target[prevLen+i-start] = v
			}
		}
	}

	return &series{
		names: prev.names,
		seq:   frame.Seq,
		ts:    slices.Concat(prev.ts, frame.Ts[start:]),
		cols:  cols,
	}
}

func printFrame(frame server.Frame, tail int) {
	kind := "append"
	if frame.Typ == server.FrameSnapshot {
		kind = "snapshot"
	}
	fmt.Printf("[%s frame] typ=%d seq=%d rows=%d cols=%d\n", kind, frame.Typ, frame.Seq, len(frame.Ts), len(frame.Cols))
	if len(frame.Ts) > 0 {
		printTable(frame.Names, frame.Ts, frame.Cols, tail)
	}
}

// printTable prints the last tail rows (0 = all), values in milliseconds.
func printTable(names []string, ts []uint32, cols [][]uint32, tail int) {
	first := 0
	if tail > 0 && len(ts) > tail {
		first = len(ts) - tail
		fmt.Printf("... %d earlier rows hidden\n", first)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	row := func(cells []string) { _, _ = fmt.Fprintln(w, strings.Join(cells, "\t")+"\t") }

	row(append([]string{"unix", "local"}, names...))
	for i := first; i < len(ts); i++ {
		cells := []string{
			strconv.FormatUint(uint64(ts[i]), 10),
			time.Unix(int64(ts[i]), 0).Format(time.DateTime),
		}
		for _, col := range cols {
			if col[i] == 0 {
				cells = append(cells, "-")
			} else {
				cells = append(cells, fmt.Sprintf("%.3f", float64(col[i])/usPerMs))
			}
		}
		row(cells)
	}
	_ = w.Flush()
}

func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
