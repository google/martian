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
	"sort"
	"sync"
	"time"

	"github.com/google/martian/v3/log"
)

// Conn wraps a net.Conn and simulates connection latency and bandwidth
// charateristics.
type Conn struct {
	Context *Context

	// Shapes represents the traffic shape map inherited from the listener.
	Shapes        *urlShapes
	GlobalBuckets map[string]*Bucket
	// LocalBuckets represents a map from the url_regexes to their dedicated buckets.
	LocalBuckets map[string]*Buckets
	Established  time.Time
	// Established is the time that the connection is established.
	DefaultBandwidth Bandwidth
	Listener         *Listener
	ReadBucket       *Bucket // Shared by listener.
	WriteBucket      *Bucket // Shared by listener.

	conn    net.Conn
	latency time.Duration
	ronce   sync.Once
	wonce   sync.Once
}

// Read reads bytes from connection into b, optionally simulating connection
// latency and throttling read throughput based on desired bandwidth
// constraints.
func (c *Conn) Read(b []byte) (int, error) {
	c.ronce.Do(c.sleepLatency)

	n, err := c.ReadBucket.FillThrottle(func(remaining int64) (int64, error) {
		max := remaining
		if l := int64(len(b)); max > l {
			max = l
		}

		n, err := c.conn.Read(b[:max])
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
func (c *Conn) ReadFrom(r io.Reader) (int64, error) {
	c.ronce.Do(c.sleepLatency)

	var total int64
	for {
		n, err := c.ReadBucket.FillThrottle(func(remaining int64) (int64, error) {
			return io.CopyN(c.conn, r, remaining)
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

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// GetWrappedConn returns the undrelying trafficshaped net.Conn.
func (c *Conn) GetWrappedConn() net.Conn {
	return c.conn
}

// WriteTo writes data to w from the connection, optionally simulating
// connection latency and throttling write throughput based on desired
// bandwidth constraints.
func (c *Conn) WriteTo(w io.Writer) (int64, error) {
	c.wonce.Do(c.sleepLatency)

	var total int64
	for {
		n, err := c.WriteBucket.FillThrottle(func(remaining int64) (int64, error) {
			return io.CopyN(w, c.conn, remaining)
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

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// CheckExistenceAndValidity checks that the current url regex is present in the map, and that
// the connection was established before the url shape map was last updated. We do not allow the
// updated url shape map to traffic shape older connections.
// Important: Assumes you have acquired the required locks and will release them youself.
func (c *Conn) CheckExistenceAndValidity(URLRegex string) bool {
	shapeStillValid := c.Shapes.LastModifiedTime.Before(c.Established)
	_, p := c.Shapes.M[URLRegex]
	return p && shapeStillValid
}

// GetCurrentThrottle uses binary search to determine if the current byte offset ('start')
// lies within a throttle interval. If so, also returns the bandwidth specified for that interval.
func (c *Conn) GetCurrentThrottle(start int64) *ThrottleContext {
	c.Shapes.RLock()
	defer c.Shapes.RUnlock()

	if !c.CheckExistenceAndValidity(c.Context.URLRegex) {
		log.Debugf("existence check failed")
		return &ThrottleContext{
			ThrottleNow: false,
		}
	}

	c.Shapes.M[c.Context.URLRegex].RLock()
	defer c.Shapes.M[c.Context.URLRegex].RUnlock()

	throttles := c.Shapes.M[c.Context.URLRegex].Shape.Throttles

	if l := len(throttles); l != 0 {
		// ind is the first index in throttles with ByteStart > start.
		// Once we get ind, we can check the previous throttle, if any,
		// to see if its ByteEnd is after 'start'.
		ind := sort.Search(len(throttles),
			func(i int) bool { return throttles[i].ByteStart > start })

		// All throttles have Bytestart > start, hence not in throttle.
		if ind == 0 {
			return &ThrottleContext{
				ThrottleNow: false,
			}
		}

		// No throttle has Bytestart > start, so check the last throttle to
		// see if it ends after 'start'. Note: the last throttle is special
		// since it can have -1 (meaning infinity) as the ByteEnd.
		if ind == l {
			if throttles[l-1].ByteEnd > start || throttles[l-1].ByteEnd == -1 {
				return &ThrottleContext{
					ThrottleNow: true,
					Bandwidth:   throttles[l-1].Bandwidth,
				}
			}
			return &ThrottleContext{
				ThrottleNow: false,
			}
		}

		// Check the previous throttle to see if it ends after 'start'.
		if throttles[ind-1].ByteEnd > start {
			return &ThrottleContext{
				ThrottleNow: true,
				Bandwidth:   throttles[ind-1].Bandwidth,
			}
		}

		return &ThrottleContext{
			ThrottleNow: false,
		}
	}

	return &ThrottleContext{
		ThrottleNow: false,
	}
}

// GetNextActionFromByte takes in a byte offset and uses binary search to determine the upcoming
// action, i.e the first action after the byte that still has a non zero count.
func (c *Conn) GetNextActionFromByte(start int64) *NextActionInfo {
	c.Shapes.RLock()
	defer c.Shapes.RUnlock()

	if !c.CheckExistenceAndValidity(c.Context.URLRegex) {
		log.Debugf("existence check failed")
		return &NextActionInfo{
			ActionNext: false,
		}
	}

	c.Shapes.M[c.Context.URLRegex].RLock()
	defer c.Shapes.M[c.Context.URLRegex].RUnlock()

	actions := c.Shapes.M[c.Context.URLRegex].Shape.Actions

	if l := len(actions); l != 0 {
		ind := sort.Search(len(actions),
			func(i int) bool { return actions[i].getByte() >= start })

		return c.GetNextActionFromIndex(int64(ind))
	}

	return &NextActionInfo{
		ActionNext: false,
	}
}

// GetNextActionFromIndex takes in an index and returns the first action after the index that
// has a non zero count, if there is one.
func (c *Conn) GetNextActionFromIndex(ind int64) *NextActionInfo {
	c.Shapes.RLock()
	defer c.Shapes.RUnlock()

	if !c.CheckExistenceAndValidity(c.Context.URLRegex) {
		return &NextActionInfo{
			ActionNext: false,
		}
	}

	c.Shapes.M[c.Context.URLRegex].RLock()
	defer c.Shapes.M[c.Context.URLRegex].RUnlock()

	actions := c.Shapes.M[c.Context.URLRegex].Shape.Actions

	if l := int64(len(actions)); l != 0 {

		for ind < l && (actions[ind].getCount() == 0) {
			ind++
		}

		if ind >= l {
			return &NextActionInfo{
				ActionNext: false,
			}
		}
		return &NextActionInfo{
			ActionNext: true,
			Index:      ind,
			ByteOffset: actions[ind].getByte(),
		}
	}
	return &NextActionInfo{
		ActionNext: false,
	}
}

// WriteDefaultBuckets writes bytes from b to the connection, optionally simulating
// connection latency and throttling write throughput based on desired
// bandwidth constraints. It uses the WriteBucket inherited from the listener.
func (c *Conn) WriteDefaultBuckets(b []byte) (int, error) {
	c.wonce.Do(c.sleepLatency)

	var total int64
	for len(b) > 0 {
		var max int64

		n, err := c.WriteBucket.FillThrottle(func(remaining int64) (int64, error) {
			max = remaining
			if l := int64(len(b)); remaining >= l {
				max = l
			}

			n, err := c.conn.Write(b[:max])
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

// Write writes bytes from b to the connection, while enforcing throttles and performing actions.
// It uses and updates the Context in the connection.
func (c *Conn) Write(b []byte) (int, error) {
	if !c.Context.Shaping {
		return c.WriteDefaultBuckets(b)
	}
	c.wonce.Do(c.sleepLatency)
	var total int64

	// Write the header if needed, without enforcing any traffic shaping, and without updating
	// ByteOffset.
	if headerToWrite := c.Context.HeaderLen - c.Context.HeaderBytesWritten; headerToWrite > 0 {
		writeAmount := min(int64(len(b)), headerToWrite)

		n, err := c.conn.Write(b[:writeAmount])

		if err != nil {
			if err != io.EOF {
				log.Errorf("trafficshape: failed write: %v", err)
			}
			return int(n), err
		}
		c.Context.HeaderBytesWritten += writeAmount
		total += writeAmount
		b = b[writeAmount:]
	}

	var amountToWrite int64

	for len(b) > 0 {
		var max int64

		// Determine the amount to be written up till the next action.
		amountToWrite = int64(len(b))
		if c.Context.NextActionInfo.ActionNext {
			amountTillNextAction := c.Context.NextActionInfo.ByteOffset - c.Context.ByteOffset
			if amountTillNextAction <= amountToWrite {
				amountToWrite = amountTillNextAction
			}
		}

		// Write into both the local and global buckets, as well as the underlying connection.
		n, err := c.Context.Buckets.WriteBucket.FillThrottleLocked(func(remaining int64) (int64, error) {
			max = min(remaining, amountToWrite)

			if max == 0 {
				return 0, nil
			}

			return c.Context.GlobalBucket.FillThrottleLocked(func(rem int64) (int64, error) {
				max = min(rem, max)
				n, err := c.conn.Write(b[:max])

				return int64(n), err
			})
		})

		if err != nil {
			if err != io.EOF {
				log.Errorf("trafficshape: failed write: %v", err)
			}
			return int(total), err
		}

		// Update the current byte offset.
		c.Context.ByteOffset += n
		total += n

		b = b[max:]

		// Check if there was an upcoming action, and that the byte offset matches the action's byte.
		if c.Context.NextActionInfo.ActionNext &&
			c.Context.ByteOffset >= c.Context.NextActionInfo.ByteOffset {
			// Note here, we check again that the url shape map is still valid and that the action still has
			// a non zero count, since that could have been modified since the last time we checked.
			ind := c.Context.NextActionInfo.Index
			c.Shapes.RLock()
			if !c.CheckExistenceAndValidity(c.Context.URLRegex) {
				c.Shapes.RUnlock()
				// Write the remaining b using default buckets, and set Shaping as false
				// so that subsequent calls to Write() also use default buckets
				// without performing any actions.
				c.Context.Shaping = false
				writeTotal, e := c.WriteDefaultBuckets(b)
				return int(total) + writeTotal, e
			}
			c.Shapes.M[c.Context.URLRegex].Lock()
			actions := c.Shapes.M[c.Context.URLRegex].Shape.Actions
			if actions[ind].getCount() != 0 {
				// Update the action count, determine the type of action and perform it.
				actions[ind].decrementCount()
				switch action := actions[ind].(type) {
				case *Halt:
					d := action.Duration
					log.Debugf("trafficshape: Sleeping for time %d ms for urlregex %s at byte offset %d",
						d, c.Context.URLRegex, c.Context.ByteOffset)
					c.Shapes.M[c.Context.URLRegex].Unlock()
					c.Shapes.RUnlock()
					time.Sleep(time.Duration(d) * time.Millisecond)
				case *CloseConnection:
					log.Infof("trafficshape: Closing connection for urlregex %s at byte offset %d",
						c.Context.URLRegex, c.Context.ByteOffset)
					c.Shapes.M[c.Context.URLRegex].Unlock()
					c.Shapes.RUnlock()
					return int(total), &ErrForceClose{message: "Forcing close connection"}
				case *ChangeBandwidth:
					bw := action.Bandwidth
					log.Infof("trafficshape: Changing connection bandwidth to %d for urlregex %s at byte offset %d",
						bw, c.Context.URLRegex, c.Context.ByteOffset)
					c.Shapes.M[c.Context.URLRegex].Unlock()
					c.Shapes.RUnlock()
					c.Context.Buckets.WriteBucket.SetCapacity(bw)
				default:
					c.Shapes.M[c.Context.URLRegex].Unlock()
					c.Shapes.RUnlock()
				}
			} else {
				c.Shapes.M[c.Context.URLRegex].Unlock()
				c.Shapes.RUnlock()
			}
			// Get the next action to be performed, if any.
			c.Context.NextActionInfo = c.GetNextActionFromIndex(ind + 1)
		}
	}
	return int(total), nil
}

func (c *Conn) sleepLatency() {
	log.Debugf("trafficshape: simulating latency: %s", c.latency)
	time.Sleep(c.latency)
}
