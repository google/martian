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
)

// DefaultBitrate represents the bitrate that will be for all url regexs for which a shape
// has not been specified.
var DefaultBitrate int64 = 500000000000 // 500Gbps (unlimited)

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

// NextActionInfo represents whether there is an upcoming action. Only if ActionNext is
// true will the Index and ByteOffset be set correctly.
type NextActionInfo struct {
	ActionNext bool
	Index      int64
	ByteOffset int64
}
