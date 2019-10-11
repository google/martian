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
	"net"
	"sync"
	"time"

	"github.com/google/martian/v3/log"
)

// DefaultBitrate represents the bitrate that will be for all url regexs for which a shape
// has not been specified.
var DefaultBitrate int64 = 500000000000 // 500Gbps (unlimited)

// ErrForceClose is an error that communicates the need to close the connection.
type ErrForceClose struct {
	message string
}

func (efc *ErrForceClose) Error() string {
	return efc.message
}

// urlShape contains a rw lock protected shape of a url_regex.
type urlShape struct {
	sync.RWMutex
	Shape *Shape
}

// urlShapes contains a rw lock protected map of url regexs to their URLShapes.
type urlShapes struct {
	sync.RWMutex
	M                map[string]*urlShape
	LastModifiedTime time.Time
}

// Buckets contains the read and write buckets for a url_regex.
type Buckets struct {
	ReadBucket  *Bucket
	WriteBucket *Bucket
}

// NewBuckets returns a *Buckets with the specified up and down bandwidths.
func NewBuckets(up int64, down int64) *Buckets {
	return &Buckets{
		ReadBucket:  NewBucket(up, time.Second),
		WriteBucket: NewBucket(down, time.Second),
	}
}

// ThrottleContext represents whether we are currently in a throttle interval for a particular
// url_regex. If ThrottleNow is true, only then will the current throttle 'Bandwidth' be set
// correctly.
type ThrottleContext struct {
	ThrottleNow bool
	Bandwidth   int64
}

// NextActionInfo represents whether there is an upcoming action. Only if ActionNext is true will the
// Index and ByteOffset be set correctly.
type NextActionInfo struct {
	ActionNext bool
	Index      int64
	ByteOffset int64
}

// Context represents the current information that is needed while writing back to the client.
// Only if Shaping is true, that is we are currently writing back a response that matches a certain
// url_regex will the other values be set correctly. If so, the Buckets represent the buckets
// to be used for the current url_regex. NextActionInfo tells us whether there is an upcoming action
// that needs to be performed, and ThrottleContext tells us whether we are currently in a throttle
// interval (according to the RangeStart). Note, the ThrottleContext is only used once in the start
// to determine the beginning bandwidth. It need not be updated after that. This
// is because the subsequent throttles are captured in the upcoming ChangeBandwidth actions.
// Byte Offset represents the absolute byte offset of response data that we are currently writing back.
// It does not account for the header data.
type Context struct {
	Shaping            bool
	RangeStart         int64
	URLRegex           string
	Buckets            *Buckets
	GlobalBucket       *Bucket
	ThrottleContext    *ThrottleContext
	NextActionInfo     *NextActionInfo
	ByteOffset         int64
	HeaderLen          int64
	HeaderBytesWritten int64
}

// Listener wraps a net.Listener and simulates connection latency and bandwidth
// constraints.
type Listener struct {
	net.Listener

	ReadBucket  *Bucket
	WriteBucket *Bucket

	mu            sync.RWMutex
	latency       time.Duration
	GlobalBuckets map[string]*Bucket
	Shapes        *urlShapes
	defaults      *Default
}

// NewListener returns a new bandwidth constrained listener. Defaults to
// DefaultBitrate (uncapped).
func NewListener(l net.Listener) *Listener {
	return &Listener{
		Listener:      l,
		ReadBucket:    NewBucket(DefaultBitrate/8, time.Second),
		WriteBucket:   NewBucket(DefaultBitrate/8, time.Second),
		Shapes:        &urlShapes{M: make(map[string]*urlShape)},
		GlobalBuckets: make(map[string]*Bucket),
		defaults: &Default{
			Bandwidth: Bandwidth{
				Up:   DefaultBitrate / 8,
				Down: DefaultBitrate / 8,
			},
			Latency: 0,
		},
	}
}

// ReadBitrate returns the bitrate in bits per second for reads.
func (l *Listener) ReadBitrate() int64 {
	return l.ReadBucket.Capacity() * 8
}

// SetReadBitrate sets the bitrate in bits per second for reads.
func (l *Listener) SetReadBitrate(bitrate int64) {
	l.ReadBucket.SetCapacity(bitrate / 8)
}

// WriteBitrate returns the bitrate in bits per second for writes.
func (l *Listener) WriteBitrate() int64 {
	return l.WriteBucket.Capacity() * 8
}

// SetWriteBitrate sets the bitrate in bits per second for writes.
func (l *Listener) SetWriteBitrate(bitrate int64) {
	l.WriteBucket.SetCapacity(bitrate / 8)
}

// SetDefaults sets the default traffic shaping parameters for the listener.
func (l *Listener) SetDefaults(defaults *Default) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.defaults = defaults
}

// Defaults returns the default traffic shaping parameters for the listener.
func (l *Listener) Defaults() *Default {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.defaults
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

// GetTrafficShapedConn takes in a normal connection and returns a traffic shaped connection.
func (l *Listener) GetTrafficShapedConn(oc net.Conn) *Conn {
	if tsconn, ok := oc.(*Conn); ok {
		return tsconn
	}
	urlbuckets := make(map[string]*Buckets)
	globalurlbuckets := make(map[string]*Bucket)

	l.Shapes.RLock()
	defaults := l.Defaults()
	latency := l.Latency()
	defaultBandwidth := defaults.Bandwidth
	for regex, shape := range l.Shapes.M {
		// It should be ok to not acquire the read lock on shape, since WriteBucket is never mutated.
		globalurlbuckets[regex] = shape.Shape.WriteBucket
		urlbuckets[regex] = NewBuckets(DefaultBitrate/8, shape.Shape.MaxBandwidth)
	}

	l.Shapes.RUnlock()

	curinfo := &Context{}

	lc := &Conn{
		conn:             oc,
		latency:          latency,
		ReadBucket:       l.ReadBucket,
		WriteBucket:      l.WriteBucket,
		Shapes:           l.Shapes,
		GlobalBuckets:    globalurlbuckets,
		LocalBuckets:     urlbuckets,
		Context:          curinfo,
		Established:      time.Now(),
		DefaultBandwidth: defaultBandwidth,
		Listener:         l,
	}
	return lc
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
	return l.GetTrafficShapedConn(oc), nil
}

// Close closes the read and write buckets along with the underlying listener.
func (l *Listener) Close() error {
	defer log.Debugf("trafficshape: closed read/write buckets and connection")

	l.ReadBucket.Close()
	l.WriteBucket.Close()

	return l.Listener.Close()
}
