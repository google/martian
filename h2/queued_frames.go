package h2

import (
	"bytes"
	"fmt"

	"golang.org/x/net/http2"
)

// queuedFrame stores frames that belong to a stream and need to be kept in order. The need for
// this stems from flow control needed in the context of gRPC. Since a gRPC message can be split
// over multiple DATA frames, the proxy needs to buffer such frames so they can be reassembled
// into messages and edited before being forwarded.
//
// Note that the proxy does man-in-the-middle flow control independently to each endpoint instead
// of forwarding endpoint flow-control messages to each other directly. This is necessary because
// multiple DATA frames need to be captured before they can be forwarded. While the data frames are
// being held in the proxy, the destination of those frames cannot see them to send WINDOW_UPDATE
// acknowledgements and the sender will stop sending data. So the proxy must emit its own
// WINDOW_UPDATEs.
//
// Example: While DATA frames are being output-buffered due to pending WINDOW_UPDATE frames from
// the destination, it's possible for the source to send subsequent HEADER frames. Those HEADER
// frames must be queued after the DATA frames for consistency with HTTP/2's total ordering of
// frames within a stream.
//
// While the example only illustrates the need for HEADER frame buffering, a similar argument
// applies to other types of stream frames. WINDOW_UPDATE is a special case that is associated
// with a stream but does not require buffering or special ordering. This is because WINDOW_UPDATEs
// are basically acknowledgements for messages coming from the peer endpoint. In other words,
// WINDOW_UPDATE frames are associated with messages being received instead of messages being sent.
// The asynchrony of receiving remote messages should allow reordering freedom.
type queuedFrame interface {
	// StreamID is the stream ID for the frame.
	StreamID() uint32

	// flowControlSize returns the size of this frame for the purposes of flow control. It is only
	// non-zero for DATA frames.
	flowControlSize() int

	// send writes the frame to the provided framer. This is not thread-safe and the caller should be
	// holding appropriate locks.
	send(*http2.Framer) error
}

type queuedDataFrame struct {
	streamID  uint32
	endStream bool
	data      []byte
}

func (f *queuedDataFrame) StreamID() uint32 {
	return f.streamID
}

func (f *queuedDataFrame) flowControlSize() int {
	return len(f.data)
}

func (f *queuedDataFrame) send(dest *http2.Framer) error {
	return dest.WriteData(f.streamID, f.endStream, f.data)
}

func (f *queuedDataFrame) String() string {
	return fmt.Sprintf("data[id=%d, endStream=%t, len=%d]", f.streamID, f.endStream, len(f.data))
}

type queuedHeaderFrame struct {
	streamID  uint32
	endStream bool
	priority  http2.PriorityParam
	chunks    [][]byte
}

func (f *queuedHeaderFrame) StreamID() uint32 {
	return f.streamID
}

func (*queuedHeaderFrame) flowControlSize() int {
	return 0
}

func (f *queuedHeaderFrame) send(dest *http2.Framer) error {
	if err := dest.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      f.streamID,
		BlockFragment: f.chunks[0],
		EndStream:     f.endStream,
		EndHeaders:    len(f.chunks) <= 1,
		PadLength:     0,
		Priority:      f.priority,
	}); err != nil {
		return fmt.Errorf("sending header %v: %w", f, err)
	}
	for i := 1; i < len(f.chunks); i++ {
		headersEnded := i == len(f.chunks)-1
		if err := dest.WriteContinuation(f.streamID, headersEnded, f.chunks[i]); err != nil {
			return fmt.Errorf("sending header continuations %v: %w", f, err)
		}
	}
	return nil
}

func (f *queuedHeaderFrame) String() string {
	var buf bytes.Buffer // strings.Builder is not available on App Engine.
	fmt.Fprintf(&buf, "header[id=%d, endStream=%t", f.streamID, f.endStream)
	fmt.Fprintf(&buf, ", priority=%v, chunk lengths=[", f.priority)
	for i, c := range f.chunks {
		if i > 0 {
			fmt.Fprintf(&buf, ",")
		}
		fmt.Fprintf(&buf, "%d", len(c))
	}
	fmt.Fprintf(&buf, "]]")
	return buf.String()
}

type queuedPushPromiseFrame struct {
	streamID  uint32
	promiseID uint32
	chunks    [][]byte
}

func (f *queuedPushPromiseFrame) StreamID() uint32 {
	return f.streamID
}

func (*queuedPushPromiseFrame) flowControlSize() int {
	return 0
}

func (f *queuedPushPromiseFrame) send(dest *http2.Framer) error {
	if err := dest.WritePushPromise(http2.PushPromiseParam{
		StreamID:      f.streamID,
		PromiseID:     f.promiseID,
		BlockFragment: f.chunks[0],
		EndHeaders:    len(f.chunks) <= 1,
		PadLength:     0,
	}); err != nil {
		return fmt.Errorf("sending push promise %v: %w", f, err)
	}
	for i := 1; i < len(f.chunks); i++ {
		headersEnded := i == len(f.chunks)-1
		if err := dest.WriteContinuation(f.streamID, headersEnded, f.chunks[i]); err != nil {
			return fmt.Errorf("sending push promise continuations %v: %w", f, err)
		}
	}
	return nil
}

func (f *queuedPushPromiseFrame) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "push promise[streamID=%d, promiseID= %d", f.streamID, f.promiseID)
	fmt.Fprintf(&buf, ", chunk lengths=[")
	for i, c := range f.chunks {
		if i > 0 {
			fmt.Fprintf(&buf, ",")
		}
		fmt.Fprintf(&buf, "%d", len(c))
	}
	fmt.Fprintf(&buf, "]]")
	return buf.String()
}

type queuedPriorityFrame struct {
	streamID uint32
	priority http2.PriorityParam
}

func (f *queuedPriorityFrame) StreamID() uint32 {
	return f.streamID
}

func (*queuedPriorityFrame) flowControlSize() int {
	return 0
}

func (f *queuedPriorityFrame) send(dest *http2.Framer) error {
	if err := dest.WritePriority(f.streamID, f.priority); err != nil {
		return fmt.Errorf("sending %v: %w", f, err)
	}
	return nil
}

func (f *queuedPriorityFrame) String() string {
	return fmt.Sprintf("priority[id=%d, priority=%v]", f.streamID, f.priority)
}

type queuedRSTStreamFrame struct {
	streamID uint32
	errCode  http2.ErrCode
}

func (f *queuedRSTStreamFrame) StreamID() uint32 {
	return f.streamID
}

func (*queuedRSTStreamFrame) flowControlSize() int {
	return 0
}

func (f *queuedRSTStreamFrame) send(dest *http2.Framer) error {
	if err := dest.WriteRSTStream(f.streamID, f.errCode); err != nil {
		return fmt.Errorf("sending %v: %w", f, err)
	}
	return nil
}

func (f *queuedRSTStreamFrame) String() string {
	return fmt.Sprintf("RSTStream[id=%d, errCode=%v]", f.streamID, f.errCode)
}
