// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trafficshape

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"
)

func TestListenerRead(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	tsl := NewListener(l)
	defer tsl.Close()

	if got := tsl.ReadBitrate(); got != defaultBitrate {
		t.Errorf("tsl.ReadBitrate(): got %d, want defaultBitrate", got)
	}
	if got := tsl.WriteBitrate(); got != defaultBitrate {
		t.Errorf("tsl.WriteBitrate(): got %d, want defaultBitrate", got)
	}

	tsl.SetReadBitrate(40) // 4 bytes per second
	if got, want := tsl.ReadBitrate(), int64(40); got != want {
		t.Errorf("tsl.ReadBitrate(): got %d, want %d", got, want)
	}

	tsl.SetWriteBitrate(40) // 4 bytes per second
	if got, want := tsl.WriteBitrate(), int64(40); got != want {
		t.Errorf("tsl.WriteBitrate(): got %d, want %d", got, want)
	}

	tsl.SetLatency(time.Second)
	if got, want := tsl.Latency(), time.Second; got != want {
		t.Errorf("tsl.Latency(): got %s, want %s", got, want)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	want := bytes.Repeat([]byte("*"), 16)

	go func() {
		// Dial the local listener.
		c, err := net.Dial("tcp", tsl.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
		defer c.Close()

		// Wait for the signal that it's okay to write to the connection; ensure
		// the test is ready to read it.
		wg.Wait()

		c.Write(want)
	}()

	tsc, err := tsl.Accept()
	if err != nil {
		t.Fatalf("tsl.Accept(): got %v, want no error", err)
	}
	defer tsc.Close()

	// Signal to the write goroutine that it may begin writing.
	wg.Done()

	start := time.Now()
	got, err := ioutil.ReadAll(tsc)
	end := time.Now()

	if err != nil {
		t.Fatalf("tsc.Read(): got %v, want no error", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("tsc.Read(): got %q, want %q", got, want)
	}

	// Breakdown of ~3s minimum:
	// 1 second for the initial latency
	// ~2-3 seconds for throttled read
	//   - 4 bytes per second with 16 bytes  total = 3 seconds (first four bytes
	//     are read immediately at the zeroth second; 0:4, 1:8, 2:12, 3:16)
	//   - the drain ticker begins before the initial start time so some of that
	//     tick time is unaccounted for in the difference; potentially up to a
	//     full second (the drain interval). For example, if the ticker is 300ms
	//     into its tick before start is calculated we will believe that the
	//     throttled read will have occurred in 2.7s. Allow for up to drain
	//     interval in skew to account for this and ensure the test does not
	//     flake.
	//
	// The test runtime should be negligible compared the latency simulation, so
	// we assume the ~3s (> 2.95s) is accounted for by throttling and latency in
	// the worst case (we read and a new tick happens immediately).
	min := 2*time.Second + 950*time.Millisecond
	max := 4*time.Second + 50*time.Millisecond
	if got := end.Sub(start); !between(got, min, max) {
		t.Errorf("tsc.Read(): took %s, want within [%s, %s]", got, min, max)
	}
}

func TestListenerWrite(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	tsl := NewListener(l)
	defer tsl.Close()

	tsl.SetReadBitrate(40)  // 4 bytes per second
	tsl.SetWriteBitrate(40) // 4 bytes per second
	tsl.SetLatency(time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	want := bytes.Repeat([]byte("*"), 16)

	var start time.Time
	var end time.Time

	go func() {
		// Dial the local listener.
		c, err := net.Dial("tcp", tsl.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
		defer c.Close()

		// Wait for the signal that it's okay to read from the connection; ensure
		// the test is ready to write to it.
		wg.Wait()

		got, err := ioutil.ReadAll(c)
		if err != nil {
			t.Fatalf("c.Read(): got %v, want no error", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("c.Read(): got %q, want %q", got, want)
		}
	}()

	tsc, err := tsl.Accept()
	if err != nil {
		t.Fatalf("tsl.Accept(): got %v, want no error", err)
	}

	// Signal to the write goroutine that it may begin writing.
	wg.Done()

	start = time.Now()
	n, err := tsc.Write(want)
	end = time.Now()

	tsc.Close()

	if err != nil {
		t.Fatalf("tsc.Write(): got %v, want no error", err)
	}
	if got, want := n, len(want); got != want {
		t.Errorf("tsc.Write(): got %d bytes, want %d bytes", got, want)
	}

	// Breakdown of ~3s minimum:
	// 1 second for the initial latency
	// ~2-3 seconds for throttled write
	//   - 4 bytes per second with 16 bytes  total = 3 seconds (first four bytes
	//     are written immediately at the zeroth second; 0:4, 1:8, 2:12, 3:16)
	//   - the drain ticker begins before the initial start time so some of that
	//     tick time is unaccounted for in the difference; potentially up to a
	//     full second (the drain interval). For example, if the ticker is 300ms
	//     into its tick before start is calculated we will believe that the
	//     throttled write will have occurred in 2.7s. Allow for up to drain
	//     interval in skew to account for this and ensure the test does not
	//     flake.
	//
	// The test runtime should be negligible compared the latency simulation, so
	// we assume the ~3s (> 2.95s) is accounted for by throttling and latency in
	// the worst case (we write and a new tick happens immediately).
	min := 2*time.Second + 950*time.Millisecond
	max := 4*time.Second + 50*time.Millisecond
	if got := end.Sub(start); !between(got, min, max) {
		t.Errorf("tsc.Write(): took %s, want within [%s, %s]", got, min, max)
	}
}

func TestListenerWriteTo(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	tsl := NewListener(l)
	defer tsl.Close()

	tsl.SetReadBitrate(40)  // 4 bytes per second
	tsl.SetWriteBitrate(40) // 4 bytes per second
	tsl.SetLatency(time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	want := bytes.Repeat([]byte("*"), 16)

	go func() {
		// Dial the local listener.
		c, err := net.Dial("tcp", tsl.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
		defer c.Close()

		// Wait for the signal that it's okay to write to the connection; ensure
		// the test is ready to read it.
		wg.Wait()

		c.Write(want)
	}()

	tsc, err := tsl.Accept()
	if err != nil {
		t.Fatalf("tsl.Accept(): got %v, want no error", err)
	}
	defer tsc.Close()

	// Signal to the write goroutine that it may begin writing.
	wg.Done()

	got := &bytes.Buffer{}

	wt, ok := tsc.(io.WriterTo)
	if !ok {
		t.Fatal("tsc.(io.WriterTo): got !ok, want ok")
	}

	start := time.Now()
	n, err := wt.WriteTo(got)
	end := time.Now()

	if err != io.EOF {
		t.Fatalf("tsc.WriteTo(): got %v, want io.EOF", err)
	}
	if got, want := n, int64(len(want)); got != want {
		t.Errorf("tsc.WriteTo(): got %d bytes, want %d bytes", got, want)
	}
	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("tsc.WriteTo(): got %q, want %q", got.Bytes(), want)
	}

	// Breakdown of ~3s minimum:
	// 1 second for the initial latency
	// ~2-3 seconds for throttled read
	//   - 4 bytes per second with 16 bytes  total = 3 seconds (first four bytes
	//     are read immediately at the zeroth second; 0:4, 1:8, 2:12, 3:16)
	//   - the drain ticker begins before the initial start time so some of that
	//     tick time is unaccounted for in the difference; potentially up to a
	//     full second (the drain interval). For example, if the ticker is 300ms
	//     into its tick before start is calculated we will believe that the
	//     throttled read will have occurred in 2.7s. Allow for up to drain
	//     interval in skew to account for this and ensure the test does not
	//     flake.
	//
	// The test runtime should be negligible compared the latency simulation, so
	// we assume the ~3s (> 2.95s) is accounted for by throttling and latency in
	// the worst case (we read and a new tick happens immediately).
	min := 2*time.Second + 950*time.Millisecond
	max := 4*time.Second + 50*time.Millisecond
	if got := end.Sub(start); !between(got, min, max) {
		t.Errorf("tsc.WriteTo(): took %s, want within [%s, %s]", got, min, max)
	}
}

func TestListenerReadFrom(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatalf("net.Listen(): got %v, want no error", err)
	}

	tsl := NewListener(l)
	defer tsl.Close()

	tsl.SetReadBitrate(40)  // 4 bytes per second
	tsl.SetWriteBitrate(40) // 4 bytes per second
	tsl.SetLatency(time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	want := bytes.Repeat([]byte("*"), 16)

	var start time.Time
	var end time.Time

	go func() {
		// Dial the local listener.
		c, err := net.Dial("tcp", tsl.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
		defer c.Close()

		// Wait for the signal that it's okay to read from the connection; ensure
		// the test is ready to write it.
		wg.Wait()

		got, err := ioutil.ReadAll(c)
		if err != nil {
			t.Fatalf("c.Read(): got %v, want no error", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("c.Read(): got %q, want %q", got, want)
		}
	}()

	tsc, err := tsl.Accept()
	if err != nil {
		t.Fatalf("tsl.Accept(): got %v, want no error", err)
	}

	// Signal to the write goroutine that it may begin writing.
	wg.Done()

	buf := bytes.NewReader(want)

	rf, ok := tsc.(io.ReaderFrom)
	if !ok {
		t.Fatal("tsc.(io.ReaderFrom): got !ok, want ok")
	}

	start = time.Now()
	n, err := rf.ReadFrom(buf)
	end = time.Now()
	tsc.Close()

	if err != nil {
		t.Fatalf("tsc.ReadFrom(): got %v, want no error", err)
	}
	if got, want := n, int64(len(want)); got != want {
		t.Errorf("tsc.ReadFrom(): got %d bytes, want %d bytes", got, want)
	}

	// Breakdown of ~3s minimum:
	// 1 second for the initial latency
	// ~2-3 seconds for throttled writes
	//   - 4 bytes per second with 16 bytes  total = 3 seconds (first four bytes
	//     are written immediately at the zeroth second; 0:4, 1:8, 2:12, 3:16)
	//   - the drain ticker begins before the initial start time so some of that
	//     tick time is unaccounted for in the difference; potentially up to a
	//     full second (the drain interval). For example, if the ticker is 300ms
	//     into its tick before start is calculated we will believe that the
	//     throttled write will have occurred in 2.7s. Allow for up to drain
	//     interval in skew to account for this and ensure the test does not
	//     flake.
	//
	// The test runtime should be negligible compared the latency simulation, so
	// we assume the ~3s (> 2.95s) is accounted for by throttling and latency in
	// the worst case (we write and a new tick happens immediately).
	min := 2*time.Second + 950*time.Millisecond
	max := 4*time.Second + 50*time.Millisecond
	if got := end.Sub(start); !between(got, min, max) {
		t.Errorf("tsc.ReadFrom(): took %s, want within [%s, %s]", got, min, max)
	}
}

func between(d, min, max time.Duration) bool {
	return d >= min && d <= max
}
