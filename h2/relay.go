// Copyright 2021 Google Inc. All rights reserved.
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

package h2

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"

	"github.com/google/martian/v3/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	// See: https://httpwg.org/specs/rfc7540.html#SettingValues
	initialMaxFrameSize       = 16384
	initialMaxHeaderTableSize = 4096

	// See: https://tools.ietf.org/html/rfc7540#section-6.9.2
	defaultInitialWindowSize = 65535

	// headersPriorityMetadataLength is the length of the priority metadata that optionally occurs at
	// the beginning of the payload of the header frame.
	//
	// See: https://tools.ietf.org/html/rfc7540#section-6.2
	headersPriorityMetadataLength = 5

	// pushPromiseMetadataLength is the length of the metadata that is part of the payload of the
	// pushPromise frame. This does not include the padding length octet, which isn't needed due to
	// the relaxed security constraints of a development proxy.
	//
	// See: https://tools.ietf.org/html/rfc7540#section-6.6
	pushPromiseMetadataLength = 4

	// outputChannelSize is the size of the output channel. Roughly, it should be large enough to
	// allow a window's worth of frames to minimize synchronization overhead.
	outputChannelSize = 15
)

// relay encapsulates a flow of h2 traffic in one direction.
type relay struct {
	dir Direction

	// srcLabel and destLabel are used only to create debugging messages.
	srcLabel, destLabel string

	src *http2.Framer

	// destMu guards writes to dest, which may occur on from either the `relayFrames` thread of
	// this relay or `peer`. `peer` writes WINDOW_UPDATE frames to this relay when it receives
	// DATA frames.
	destMu sync.Mutex
	dest   *http2.Framer

	// maxFrameSize is set by the peer relay and is accessed atomically.
	maxFrameSize uint32

	// The decoder and encoder settings can be adjusted by the peer connection so access to these
	// fields must be guarded.
	decoderMu sync.Mutex
	decoder   *hpack.Decoder

	encoderMu sync.Mutex
	encoder   *hpack.Encoder
	reencoded bytes.Buffer // handle to the output buffer of `encoder`

	// headerBuffer collects header fragments that are received across multiple frames, i.e.,
	// when there are continuation frames.
	headerBuffer      bytes.Buffer
	continuationState continuationState

	// flowMu guards access to flow-control related fields.
	flowMu               sync.Mutex
	initialWindowSize    uint32
	connectionWindowSize int // "global" connection-level window size
	// outputBuffers is output pending available window size per-stream
	outputBuffers map[uint32]*outputBuffer
	// output stores stream output that is ready to be sent over HTTP/2. It provides a way to
	// guarantee frame order without blocking on each frame being sent.
	output chan queuedFrame

	enableDebugLogs *bool

	// The following fields depend on a circular dependency between the relays in opposite directions
	// so must be set explicitly after initialization.

	// processors stores per HTTP/2 stream processors.
	processors *streamProcessors

	peer *relay // relay for traffic from the peer
}

// newRelay initializes a relay for the given direction. This performs only partial initialization
// due to circular dependency.
func newRelay(
	dir Direction,
	srcLabel, destLabel string,
	src, dest *http2.Framer,
	enableDebugLogs *bool,
) *relay {
	ret := &relay{
		dir:                  dir,
		srcLabel:             srcLabel,
		destLabel:            destLabel,
		src:                  src,
		dest:                 dest,
		maxFrameSize:         initialMaxFrameSize,
		decoder:              hpack.NewDecoder(initialMaxHeaderTableSize, nil),
		initialWindowSize:    defaultInitialWindowSize,
		connectionWindowSize: defaultInitialWindowSize,
		outputBuffers:        make(map[uint32]*outputBuffer),
		output:               make(chan queuedFrame, outputChannelSize),
		enableDebugLogs:      enableDebugLogs,
	}
	ret.encoder = hpack.NewEncoder(&ret.reencoded)

	// These limits seem to be part of the Go implementation of hpack. They exist because in a
	// production system, there must be limits on the resources requested by clients. However, this
	// is irrevelevant in a development proxy context.
	ret.decoder.SetAllowedMaxDynamicTableSize(math.MaxUint32)
	ret.encoder.SetMaxDynamicTableSizeLimit(math.MaxUint32)
	return ret
}

// relayFrames reads frames from `f.src` to `f.dest` until an error occurs or the connection closes.
func (r *relay) relayFrames(closing chan bool) error {
	errChan := make(chan error)
	go func() {
		// Delivers the strictly ordered stream output.
		for {
			select {
			case f := <-r.output:
				r.destMu.Lock()
				err := f.send(r.dest)
				r.destMu.Unlock()
				if err != nil {
					errChan <- err
					return
				}
			case <-closing:
				return
			}
		}
	}()

	// This channel is buffered to allow the frame reader to drain on cancellation.
	frameReady := make(chan struct{}, 1)
	for {
		// Reads the frame from a goroutine to make this function responsive to cancellation.
		var frame http2.Frame
		var err error
		go func() {
			frame, err = r.src.ReadFrame()
			frameReady <- struct{}{}
		}()
		select {
		case <-frameReady:
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("reading frame: %w", err)
			}
			if err := r.processFrame(frame); err != nil {
				return fmt.Errorf("processing frame: %w", err)
			}
			if *r.enableDebugLogs {
				log.Infof("%s--%v-->%s", r.srcLabel, frame, r.destLabel)
			}
		case err := <-errChan:
			return fmt.Errorf("sending frame: %w", err)
		case <-closing:
			// The reader goroutine is abandoned at this point. It completes as soon as the blocking
			// ReadFrame call completes, but could potentially leak for an unspecified duration.
			return nil
		}
	}
}

func (r *relay) processFrame(f http2.Frame) error {
	var err error
	switch f := f.(type) {
	case *http2.DataFrame:
		// The proxy's window increments as soon as it receives data. This assumes that the proxy has
		// ample resources because it is inteded for testing and development.
		if err = r.peer.sendWindowUpdates(f); err == nil {
			err = r.processor(f.StreamID).Data(f.Data(), f.StreamEnded())
		}
	case *http2.HeadersFrame:
		if !f.HeadersEnded() {
			r.headerBuffer.Reset()
			r.headerBuffer.Write(f.HeaderBlockFragment())
			r.continuationState = &headerContinuation{f.Priority}
		} else {
			var headers []hpack.HeaderField
			headers, err = r.decodeFull(f.HeaderBlockFragment())
			if err != nil {
				return fmt.Errorf("decoding header %v: %w", f, err)
			}
			err = r.processor(f.StreamID).Header(headers, f.StreamEnded(), f.Priority)
		}
	case *http2.PriorityFrame:
		err = r.processor(f.StreamID).Priority(f.PriorityParam)
	case *http2.RSTStreamFrame:
		err = r.processor(f.StreamID).RSTStream(f.ErrCode)
	case *http2.SettingsFrame:
		if f.IsAck() {
			r.destMu.Lock()
			err = r.dest.WriteSettingsAck()
			r.destMu.Unlock()
		} else {
			var settings []http2.Setting
			if err = f.ForeachSetting(func(s http2.Setting) error {
				switch s.ID {
				case http2.SettingHeaderTableSize:
					r.peer.updateTableSize(s.Val)
				case http2.SettingInitialWindowSize:
					r.peer.updateInitialWindowSize(s.Val)
				case http2.SettingMaxFrameSize:
					r.peer.updateMaxFrameSize(s.Val)
				}
				settings = append(settings, s)
				return nil
			}); err == nil {
				r.destMu.Lock()
				err = r.dest.WriteSettings(settings...)
				r.destMu.Unlock()
			}
		}
	case *http2.PushPromiseFrame:
		if !f.HeadersEnded() {
			r.headerBuffer.Reset()
			r.headerBuffer.Write(f.HeaderBlockFragment())
			r.continuationState = &pushPromiseContinuation{f.PromiseID}
		} else {
			var headers []hpack.HeaderField
			headers, err = r.decodeFull(f.HeaderBlockFragment())
			if err != nil {
				return fmt.Errorf("decoding push promise %v: %w", f, err)
			}
			err = r.processor(f.StreamID).PushPromise(f.PromiseID, headers)
		}
	case *http2.PingFrame:
		r.destMu.Lock()
		err = r.dest.WritePing(f.IsAck(), f.Data)
		r.destMu.Unlock()
	case *http2.GoAwayFrame:
		r.destMu.Lock()
		err = r.dest.WriteGoAway(f.LastStreamID, f.ErrCode, f.DebugData())
		r.destMu.Unlock()
	case *http2.WindowUpdateFrame:
		r.peer.updateWindow(f)
	case *http2.ContinuationFrame:
		r.headerBuffer.Write(f.HeaderBlockFragment())
		if f.HeadersEnded() {
			var headers []hpack.HeaderField
			headers, err = r.decodeFull(r.headerBuffer.Bytes())
			if err != nil {
				return fmt.Errorf("decoding headers for continuation %v: %w", f, err)
			}
			err = r.continuationState.complete(r.processor(f.StreamID), headers)
		}
	default:
		err = errors.New("unrecognized frame type")
	}
	return err
}

func (r *relay) processor(id uint32) Processor {
	return r.processors.Get(id, r.dir)
}

func (r *relay) updateTableSize(v uint32) {
	r.decoderMu.Lock()
	r.decoder.SetMaxDynamicTableSize(v)
	r.decoderMu.Unlock()

	r.encoderMu.Lock()
	r.encoder.SetMaxDynamicTableSize(v)
	r.encoderMu.Unlock()
}

func (r *relay) updateMaxFrameSize(v uint32) {
	atomic.StoreUint32(&r.maxFrameSize, v)
}

// updateInitialWindowSize updates the initial window size and updates all stream windows based on
// the difference. Note that this should not include the connection window.
// See: https://tools.ietf.org/html/rfc7540#section-6.9.2
//
// This is called by `peer`, so requires a thread-safe implementation.
func (r *relay) updateInitialWindowSize(v uint32) {
	r.flowMu.Lock()
	delta := int(v) - int(r.initialWindowSize)
	r.initialWindowSize = v
	for _, w := range r.outputBuffers {
		w.windowSize += delta
	}
	r.flowMu.Unlock()
	// Since all the stream windows may be impacted, all the queues need to be checked for newly
	// eligible frames.
	r.sendQueuedFramesUnderWindowSize()
}

// updateWindow updates the specified window size and may result in the sending of data frames.
func (r *relay) updateWindow(f *http2.WindowUpdateFrame) {
	if f.StreamID == 0 {
		// A stream ID of 0 means to updating the global connection window size. This may cause any
		// queued frame belonging to any frame to become eligible for sending.
		r.flowMu.Lock()
		r.connectionWindowSize += int(f.Increment)
		r.flowMu.Unlock()
		r.sendQueuedFramesUnderWindowSize()
	}

	r.flowMu.Lock()
	w := r.outputBuffer(f.StreamID)
	w.windowSize += int(f.Increment)
	w.emitEligibleFrames(r.output, &r.connectionWindowSize)
	r.flowMu.Unlock()
}

func (r *relay) data(id uint32, data []byte, streamEnded bool) error {
	// This implementation only allows `WriteData` without padding. Padding is used to improve the
	// security against attacks like CRIME, but this isn't relevant for a development proxy.
	//
	// If padding were allowed, this length would need to vary depending on whether the padding
	// length octet is present.
	maxPayloadLength := atomic.LoadUint32(&r.maxFrameSize)

	r.flowMu.Lock()
	w := r.outputBuffer(id)
	r.flowMu.Unlock()
	// If data is larger than what would be permitted at the current max frame size setting, the data
	// is split across multiple frames.
	for {
		nextPayloadLength := uint32(len(data))
		if nextPayloadLength > maxPayloadLength {
			nextPayloadLength = maxPayloadLength
		}
		nextPayload := make([]byte, nextPayloadLength)
		copy(nextPayload, data)
		data = data[nextPayloadLength:]
		f := &queuedDataFrame{id, streamEnded && len(data) == 0, nextPayload}

		r.flowMu.Lock()
		w.enqueue(f)
		w.emitEligibleFrames(r.output, &r.connectionWindowSize)
		r.flowMu.Unlock()

		// Some protocols send empty data frames with END_STREAM so the check is done here at the end
		// of the loop instead of at the beginning of the loop.
		if len(data) == 0 {
			break
		}
	}
	return nil
}

func (r *relay) header(
	id uint32,
	headers []hpack.HeaderField,
	streamEnded bool,
	priority http2.PriorityParam,
) error {
	encoded, err := r.encodeFull(headers)
	if err != nil {
		return fmt.Errorf("encoding headers %v: %w", headers, err)
	}

	maxPayloadLength := atomic.LoadUint32(&r.maxFrameSize)
	// Padding is not implemented because the extra security is not needed for a development proxy.
	// If it were used, a single padding length octet should be deducted from the max header fragment
	// length.
	maxHeaderFragmentLength := maxPayloadLength
	if !priority.IsZero() {
		maxHeaderFragmentLength -= headersPriorityMetadataLength
	}
	chunks := splitIntoChunks(int(maxHeaderFragmentLength), int(maxPayloadLength), encoded)

	r.enqueueFrame(&queuedHeaderFrame{
		streamID:  id,
		endStream: streamEnded,
		priority:  priority,
		chunks:    chunks,
	})
	return nil
}

func (r *relay) priority(id uint32, priority http2.PriorityParam) {
	r.enqueueFrame(&queuedPriorityFrame{
		streamID: id,
		priority: priority,
	})
}

func (r *relay) rstStream(id uint32, errCode http2.ErrCode) {
	r.enqueueFrame(&queuedRSTStreamFrame{
		streamID: id,
		errCode:  errCode,
	})
}

func (r *relay) pushPromise(id, promiseID uint32, headers []hpack.HeaderField) error {
	encoded, err := r.encodeFull(headers)
	if err != nil {
		return fmt.Errorf("encoding push promise headers %v: %w", headers, err)
	}

	maxPayloadLength := atomic.LoadUint32(&r.maxFrameSize)
	maxHeaderFragmentLength := maxPayloadLength - pushPromiseMetadataLength
	chunks := splitIntoChunks(int(maxHeaderFragmentLength), int(maxPayloadLength), encoded)

	r.enqueueFrame(&queuedPushPromiseFrame{
		streamID:  id,
		promiseID: promiseID,
		chunks:    chunks,
	})
	return nil
}

func (r *relay) enqueueFrame(f queuedFrame) {
	// The frame is first added to the appropriate stream.
	r.flowMu.Lock()
	w := r.outputBuffer(f.StreamID())
	w.enqueue(f)
	w.emitEligibleFrames(r.output, &r.connectionWindowSize)
	r.flowMu.Unlock()
}

func (r *relay) sendQueuedFramesUnderWindowSize() {
	r.flowMu.Lock()
	for _, w := range r.outputBuffers {
		w.emitEligibleFrames(r.output, &r.connectionWindowSize)
	}
	r.flowMu.Unlock()
}

// outputBuffer returns the outputBuffer instance for the given stream, creating one if needed.
//
// This method is not thread-safe. The caller should be holding `flowMu`.
func (r *relay) outputBuffer(streamID uint32) *outputBuffer {
	w, ok := r.outputBuffers[streamID]
	if !ok {
		w = &outputBuffer{
			windowSize: int(r.initialWindowSize),
		}
		r.outputBuffers[streamID] = w
	}
	return w
}

// sendWindowUpdates sends WINDOW_UPDATE frames effectively acknowledging consumption of the
// given data frame.
func (r *relay) sendWindowUpdates(f *http2.DataFrame) error {
	if len(f.Data()) <= 0 {
		return nil
	}
	r.destMu.Lock()
	defer r.destMu.Unlock()
	// First updates the connection level window.
	if err := r.dest.WriteWindowUpdate(0, uint32(len(f.Data()))); err != nil {
		return err
	}
	// Next updates the stream specific window.
	return r.dest.WriteWindowUpdate(f.StreamID, uint32(len(f.Data())))
}

func (r *relay) decodeFull(data []byte) ([]hpack.HeaderField, error) {
	r.decoderMu.Lock()
	defer r.decoderMu.Unlock()
	return r.decoder.DecodeFull(data)
}

func (r *relay) encodeFull(headers []hpack.HeaderField) ([]byte, error) {
	r.encoderMu.Lock()
	defer r.encoderMu.Unlock()

	r.reencoded.Reset()
	var buf bytes.Buffer
	for _, h := range headers {
		if *r.enableDebugLogs {
			if h.Name == "content-type" && h.Value == "application/grpc" {
				fmt.Fprintf(&buf, "  \u001b[1;36m%v\u001b[0m\n", h)
			} else {
				fmt.Fprintf(&buf, "  %v\n", h)
			}
		}
		if err := r.encoder.WriteField(h); err != nil {
			return nil, fmt.Errorf("reencoding header field %v in %v: %w", h, headers, err)
		}
	}
	if *r.enableDebugLogs {
		log.Infof("sending headers %s -> %s:\n%s", r.srcLabel, r.destLabel, buf.Bytes())
	}
	return r.reencoded.Bytes(), nil
}

// outputBuffer stores enqueued output frames for a given stream.
type outputBuffer struct {
	// windowSize indicates how much data the receiver is ready to process.
	windowSize int
	queue      list.List // contains queuedFrame elements
}

// emitEligibleFrames emits frames that would fit under both the stream window size and the
// given connection window size. It updates the given connectionWindowSize if applicable.
//
// This is not thread-safe. The caller should be holding `relay.flowMu`.
func (w *outputBuffer) emitEligibleFrames(output chan queuedFrame, connectionWindowSize *int) {
	for e := w.queue.Front(); e != nil; {
		f := e.Value.(queuedFrame)
		if f.flowControlSize() > *connectionWindowSize || f.flowControlSize() > w.windowSize {
			break
		}
		output <- f

		*connectionWindowSize -= f.flowControlSize()
		w.windowSize -= f.flowControlSize()

		next := e.Next()
		w.queue.Remove(e)
		e = next
	}
}

// enqueue adds the frame to this stream output. This is not thread-safe. The caller must hold
// relay.flowMu.
func (w *outputBuffer) enqueue(f queuedFrame) {
	w.queue.PushBack(f)
}

// continuationState holds the context needed to interpret CONTINUATION frames, specifically whether
// the parents were HEADERS or PUSH_PROMISE frames.
type continuationState interface {
	complete(s Processor, headers []hpack.HeaderField) error
}

type headerContinuation struct {
	priority http2.PriorityParam
}

func (h *headerContinuation) complete(s Processor, headers []hpack.HeaderField) error {
	return s.Header(headers, true, h.priority)
}

type pushPromiseContinuation struct {
	promiseID uint32
}

func (p *pushPromiseContinuation) complete(s Processor, headers []hpack.HeaderField) error {
	return s.PushPromise(p.promiseID, headers)
}

// splitIntoChunks splits header payloads into chunks that respect frame size limits.
func splitIntoChunks(firstChunkMax, continuationMax int, data []byte) [][]byte {
	var chunks [][]byte

	firstChunkLength := len(data)
	if firstChunkLength > firstChunkMax {
		firstChunkLength = firstChunkMax
	}
	buf := make([]byte, firstChunkLength)
	copy(buf, data[:firstChunkLength])
	chunks = append(chunks, buf)
	remaining := data[firstChunkLength:]
	for len(remaining) > 0 {
		nextChunkLength := len(remaining)
		if nextChunkLength > continuationMax {
			nextChunkLength = continuationMax
		}
		buf = make([]byte, nextChunkLength)
		copy(buf, remaining[:nextChunkLength])
		chunks = append(chunks, buf)
		remaining = remaining[nextChunkLength:]
	}
	return chunks
}
