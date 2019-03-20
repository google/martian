package player

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/google/martian/log"
)

// BodySegment is a segment of an HTTP response body for a specific byte range.
type BodySegment struct {
	start  int
	end    int
	offset int
	n      int

	segment io.ReadCloser
	mu      sync.Mutex
}

func (bs *BodySegment) SetRange(offset int, n int) error {
	bs.offset = offset
	bs.n = n

	// do a range check return an error if range exceedes possible len
	if bs.start+offset+(n-1) > bs.end {
		return fmt.Errorf("start:%d + offset:%d + (n:%d + 1) exceeds last byte index %d", bs.start, offset, n, bs.end)
	}

	return nil
}

func (bs *BodySegment) Read(b []byte) (int, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	log.Infof("** reading from segment [%d - %d] **", bs.start, bs.end)

	log.Infof("offset: %d", bs.offset)

	var buf bytes.Buffer

	// read bytes to be skipped over into b
	if bs.offset > 0 {
		skpb := make([]byte, bs.offset)
		if _, err := io.ReadAtLeast(bs.segment, skpb, bs.offset); err != nil && err != io.EOF {
			return 0, err
		}

		log.Infof("skipped %d bytes:", len(skpb))

		if _, err := buf.Write(skpb); err != nil && err != io.EOF {
			return 0, err
		}
	}
	log.Infof("read buf len    : %d", buf.Len())

	// tee the bytes read from bs.segment onto buf
	tee := io.TeeReader(bs.segment, &buf)

	log.Infof("io.LimitReader(tee, %d)", bs.n)
	limr := io.LimitReader(tee, int64(bs.n))
	// return lerr to bubble up EOF
	n, lerr := limr.Read(b)
	if lerr != nil && lerr != io.EOF {
		return n, lerr
	}

	log.Infof("n          : %d", n)
	log.Infof("read buf len    : %d", buf.Len())

	// read the rest of the bytes in
	if _, err := buf.ReadFrom(bs.segment); err != nil && err != io.EOF {
		return 0, err
	}

	log.Infof("read buf len    : %d", buf.Len())

	// replace read bytes - this causes an infinite loop - fun
	// bs.segment = ioutil.NopCloser(&buf)

	return n, lerr
}

// ByStartIndex implements sort.Interface for []BodySegment based on
// the start byte index.
type ByStartIndex []*BodySegment

func (bsi ByStartIndex) Len() int      { return len(bsi) }
func (bsi ByStartIndex) Swap(i, j int) { bsi[i], bsi[j] = bsi[j], bsi[i] }
func (bsi ByStartIndex) Less(i, j int) bool {
	if bsi[i].start == bsi[j].start {
		return bsi[i].end < bsi[j].end
	}
	return bsi[i].start < bsi[j].start
}

// RangedBodyReader is an io.Reader that provides a single reader for multiple response
// bodies for the same resource with different byte ranges. Start is the first byte index
// to be read and End is the last index to be read (inclusive).
type RangedBodyReader struct {
	Start      int
	End        int
	BodyReader io.Reader

	mu       sync.Mutex
	segments []*BodySegment
	parsed   bool
}

// NewRangedBodyReader returns an io.ReadCloser that responds to ranged request headers based
// on the aggregate of the response bodies in resps.
func NewRangedBodyReader(rurl *url.URL, resps []*http.Response) (*RangedBodyReader, error) {
	rresps := make([]*http.Response, 0)
	for _, res := range resps {
		if res.Request.URL.String() != rurl.String() {
			continue
		}
		rresps = append(rresps, res)
	}

	if len(rresps) == 0 {
		return nil, fmt.Errorf("len(resps): length of ranged responses is zero; it should be greater than zero")
	}

	rbr := &RangedBodyReader{}

	for _, res := range rresps {
		cr := res.Header.Get("Content-Range")
		if cr == "" {
			return nil, fmt.Errorf("Header.Get(%q): got no value, want a properly formatted Content-Range", "Content-Range")
		}

		start, end, _, err := parseContentRangeHeader(cr)
		if err != nil {
			return nil, fmt.Errorf("parseContentRangeHeader(%q): error: %s", cr, err)
		}

		var buf bytes.Buffer
		tee := io.TeeReader(res.Body, &buf)

		bseg := &BodySegment{start: start, end: end, segment: ioutil.NopCloser(tee)}
		rbr.segments = append(rbr.segments, bseg)
		sort.Sort(ByStartIndex(rbr.segments))
		rbr.parsed = false

		// race for use of res.Body before seg.Read is called?
		res.Body = ioutil.NopCloser(&buf)
	}

	return rbr, nil
}

// ParseRange constructs the underlying reader by iterating over rbr.segments
// and assembiling a single reader. Parse must be called prior to Read.
func (rbr *RangedBodyReader) ParseRange(start, end int) error {
	rbr.mu.Lock()
	defer rbr.mu.Unlock()

	rbr.Start = start
	rbr.End = end

	// check for non-continuous ranges
	rstart, rend := 0, 0
	for i, s := range rbr.segments {
		if i == 0 {
			rstart, rend = s.start, s.end
		}

		if !(s.start >= rstart && s.start <= rend) {
			return fmt.Errorf("segments must be contiguous or overlapping. Gap between indexes %d and %d", rend, s.start)
		}

		if s.end > rend {
			rend = s.end
		}
	}

	log.Infof("total range: %d - %d", rstart, rend)
	log.Infof("  rbr range: %d - %d", rbr.Start, rbr.End)

	if rbr.segments[0].start != 0 {
		return fmt.Errorf("segments must begin with index 0")
	}

	// set up the multireader
	rdrs := make([]io.Reader, 0)
	lastParsed := 0
	for i, seg := range rbr.segments {
		if (rbr.Start > seg.end) || (rbr.End < seg.start) {
			log.Infof("segment %d - %d is NOT a part of the response range %d - %d", seg.start, seg.end, rbr.Start, rbr.End)
			continue
		}

		if lastParsed >= seg.end {
			log.Infof("beyond segment %d - %d is NOT a part of the response range %d - %d", seg.start, seg.end, rbr.Start, rbr.End)
			continue
		}

		log.Infof("segment %d - %d is a part of the response range %d - %d", seg.start, seg.end, rbr.Start, rbr.End)

		if lastParsed < seg.start {
			log.Infof("lastParsed: %d < seg.start: %d", lastParsed, seg.start)
			lastParsed = seg.start - 1
		}

		if lastParsed < rbr.Start {
			log.Infof("lastParsed: %d < rbr.Start: %d", lastParsed, seg.start)
			lastParsed = rbr.Start - 1
		}

		log.Infof("  lastParsed: %d", lastParsed)

		offset := lastParsed - seg.start + 1
		if offset < 0 {
			log.Infof("offset was < 0, setting to rbr.Start(%d) - seg.start(%d)", rbr.Start, seg.start)
			offset = rbr.Start - seg.start
		}
		if rbr.Start == seg.start {
			log.Infof("rbr.Start(%d) == seg.start(%d), offset = 0", rbr.Start, seg.start)
			offset = 0
		}

		log.Infof("offset: %d", offset)

		log.Infof("n = rbr.End(%d) - seg.start(%d) - offset(%d) + 1", rbr.End, seg.start, offset)
		n := rbr.End - seg.start - offset + 1
		if seg.end < rbr.End {
			n = seg.end - seg.start - offset + 1
		}

		log.Infof("     n: %d", n)

		lastParsed = seg.start + offset + n - 1
		log.Infof("setting lastParsed: %d", lastParsed)

		log.Infof("setting range for segment %d: offset: %d n: %d", i, offset, n)
		if err := seg.SetRange(offset, n); err != nil {
			return fmt.Errorf("seg.SetRange(%d, %d): error: %s", offset, n, err)
		}
		rdrs = append(rdrs, seg)

		continue
	}

	log.Infof("io.Multireader(rdrs...): len(rdrs): %d", len(rdrs))
	rbr.BodyReader = io.MultiReader(rdrs...)

	rbr.parsed = true
	return nil
}

func (rbr *RangedBodyReader) Read(b []byte) (int, error) {
	rbr.mu.Lock()
	defer rbr.mu.Unlock()

	if !rbr.parsed {
		return 0, fmt.Errorf("must call ParseRange prior to Read")
	}

	return rbr.BodyReader.Read(b)
}

// Close calls Close on each rbr.segment.
func (rbr *RangedBodyReader) Close() error {
	rbr.mu.Lock()
	defer rbr.mu.Unlock()

	for _, seg := range rbr.segments {
		seg.segment.Close()
	}

	return nil
}

func parseContentRangeHeader(value string) (int, int, int, error) {
	value = strings.TrimPrefix(value, "bytes ")

	crs := strings.Split(value, "/")
	if len(crs) != 2 {
		return 0, 0, 0, fmt.Errorf("Content-Range: %q improperly formatted without %q, want a value of the form %q", value, "/", "bytes n-x/y")
	}

	length, err := strconv.Atoi(crs[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("strconv.Atoi(%q): error: %s", crs[1], err)
	}

	rng := strings.Split(crs[0], "-")
	if len(rng) != 2 {
		return 0, 0, 0, fmt.Errorf("Content-Range: %q improperly formatted without %q, want a value of the form %q", value, "-", "bytes n-x/y")
	}

	start, err := strconv.Atoi(rng[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("strconv.Atoi(%q): error: %s", rng[0], err)
	}

	end, err := strconv.Atoi(rng[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("strconv.Atoi(%q): error: %s", rng[1], err)
	}

	return start, end, length, nil
}
