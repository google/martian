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
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/martian/log"
)

var defaultBitrate int64 = 500000000000 // 500Gbps (unlimited)

// Listener wraps a net.Listener and simulates connection latency and bandwidth
// constraints.
type Listener struct {
	net.Listener

	rb *Bucket
	wb *Bucket

	mu      sync.Mutex
	latency time.Duration
}

type conn struct {
	net.Conn

	rb      *Bucket // Shared by listener.
	wb      *Bucket // Shared by listener.
	latency time.Duration
	ronce   sync.Once
	wonce   sync.Once
}

// NewListener returns a new bandwidth constrained listener. Defaults to
// defaultBitrate (uncapped).
func NewListener(l net.Listener) *Listener {
	return &Listener{
		Listener: l,
		rb:       NewBucket(defaultBitrate/8, time.Second),
		wb:       NewBucket(defaultBitrate/8, time.Second),
	}
}

// ReadBitrate returns the bitrate in bits per second for reads.
func (l *Listener) ReadBitrate() int64 {
	return l.rb.Capacity() * 8
}

// SetReadBitrate sets the bitrate in bits per second for reads.
func (l *Listener) SetReadBitrate(bitrate int64) {
	l.rb.SetCapacity(bitrate / 8)
}

// WriteBitrate returns the bitrate in bits per second for writes.
func (l *Listener) WriteBitrate() int64 {
	return l.wb.Capacity() * 8
}

// SetWriteBitrate sets the bitrate in bits per second for writes.
func (l *Listener) SetWriteBitrate(bitrate int64) {
	l.wb.SetCapacity(bitrate / 8)
}

// Latency returns the latency for connections.
func (l *Listener) Latency() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.latency
}

// SetLatency sets the initial latency for connections.
func (l *Listener) SetLatency(latency time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.latency = latency
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	oc, err := l.Listener.Accept()
	if err != nil {
		log.Errorf("trafficshape: failed accepting connection: %v", err)
		return nil, err
	}

	if tconn, ok := oc.(*net.TCPConn); ok {
		log.Debugf("trafficshape: setting keep-alive for TCP connection")
		tconn.SetKeepAlive(true)
		tconn.SetKeepAlivePeriod(3 * time.Minute)
	}

	lc := &conn{
		Conn:    oc,
		latency: l.Latency(),
		rb:      l.rb,
		wb:      l.wb,
	}

	return lc, nil
}

// Close closes the read and write buckets along with the underlying listener.
func (l *Listener) Close() error {
	defer log.Debugf("trafficshape: closed read/write buckets and connection")

	l.rb.Close()
	l.wb.Close()

	return l.Listener.Close()
}

// Read reads bytes from connection into b, optionally simulating connection
// latency and throttling read throughput based on desired bandwidth
// constraints.
func (c *conn) Read(b []byte) (int, error) {
	c.ronce.Do(c.sleepLatency)

	n, err := c.rb.FillThrottle(func(remaining int64) (int64, error) {
		max := remaining
		if l := int64(len(b)); max > l {
			max = l
		}

		n, err := c.Conn.Read(b[:max])
		return int64(n), err
	})
	if err != nil && err != io.EOF {
		log.Errorf("trafficshape: error on throttled read: %v", err)
	}

	return int(n), err
}

// ReadFrom reads data from r until EOF or error, optionally simulating
// connection latency and throttling read throughput based on desired bandwidth
// constraints.
func (c *conn) ReadFrom(r io.Reader) (int64, error) {
	c.ronce.Do(c.sleepLatency)

	var total int64
	for {
		n, err := c.rb.FillThrottle(func(remaining int64) (int64, error) {
			return io.CopyN(c.Conn, r, remaining)
		})

		total += n

		if err == io.EOF {
			log.Debugf("trafficshape: exhausted reader successfully")
			return total, nil
		} else if err != nil {
			log.Errorf("trafficshape: failed copying from reader: %v", err)
			return total, err
		}
	}
}

// WriteTo writes data to w from the connection, optionally simulating
// connection latency and throttling write throughput based on desired
// bandwidth constraints.
func (c *conn) WriteTo(w io.Writer) (int64, error) {
	c.wonce.Do(c.sleepLatency)

	var total int64
	for {
		n, err := c.wb.FillThrottle(func(remaining int64) (int64, error) {
			return io.CopyN(w, c.Conn, remaining)
		})

		total += n

		if err != nil {
			if err != io.EOF {
				log.Errorf("trafficshape: failed copying to writer: %v", err)
			}
			return total, err
		}
	}
}

// Writes writes bytes from b to the connection, optionally simulating
// connection latency and throttling write throughput based on desired
// bandwidth constraints.
func (c *conn) Write(b []byte) (int, error) {
	c.wonce.Do(c.sleepLatency)

	var total int64
	for len(b) > 0 {
		var max int64

		n, err := c.wb.FillThrottle(func(remaining int64) (int64, error) {
			max = remaining
			if l := int64(len(b)); remaining >= l {
				max = l
			}

			n, err := c.Conn.Write(b[:max])
			return int64(n), err
		})

		total += n

		if err != nil {
			if err != io.EOF {
				log.Errorf("trafficshape: failed write: %v", err)
			}
			return int(total), err
		}

		b = b[max:]
	}

	return int(total), nil
}

func (c *conn) sleepLatency() {
	log.Debugf("trafficshape: simulating latency: %s", c.latency)
	time.Sleep(c.latency)
}
