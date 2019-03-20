package player

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/martian"
	"github.com/google/martian/proxyutil"
)

func TestRangedBodyReader(t *testing.T) {
	// 1000 bytes of lorem ipsum
	bwnt := ([]byte)("Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. N")

	//bwnt := make([]byte, len(b))
	//copy(bwnt, b)

	var rngs = []struct {
		start int
		end   int
	}{
		{0, 8},
		{5, 12},
		{12, 20},
		{10, 300},
		{299, 999},
	}

	resps := make([]*http.Response, 0)

	for _, rng := range rngs {
		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("http.NewRequest(): got %v, want no error", err)
		}

		_, removeCtx, err := martian.TestContext(req, nil, nil)
		if err != nil {
			t.Fatalf("martian.TestContext(): got %v, want no error", err)
		}
		defer removeCtx()

		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rng.start, rng.end))

		rbdy := bytes.NewBuffer(bwnt[rng.start : rng.end+1])
		res := proxyutil.NewResponse(http.StatusPartialContent, rbdy, req)
		res.Header.Set("Content-Length", string(rbdy.Len()))
		res.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rng.start, rng.end, len(bwnt)))
		res.Header.Set("Accept-Ranges", "bytes")

		resps = append(resps, res)

		removeCtx()
	}

	// make several range requests and make sure that the bytes match
	var trngs = []struct {
		start int
		end   int
	}{
		{0, 20},
		{0, 301},
		//{50, 150},
		//{29, 300},
		//{30, 50},
	}

	for i, trng := range trngs {
		t.Run(fmt.Sprintf("%d. request byte range %d-%d of %d bytes", i, trng.start, trng.end, 1000), func(t *testing.T) {
			treq, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatalf("http.NewRequest(): got %v, want no error", err)
			}
			treq.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", trng.start, trng.end))

			_, removeCtx, err := martian.TestContext(treq, nil, nil)
			if err != nil {
				t.Fatalf("martian.TestContext(): got %v, want no error", err)
			}
			defer removeCtx()

			rbr, err := NewRangedBodyReader(&url.URL{Scheme: "http", Host: "example.com"}, resps)
			if err != nil {
				t.Fatalf("NewRangedBodyReader(): error: %s", err)
			}

			if err := rbr.ParseRange(trng.start, trng.end); err != nil {
				t.Fatalf("rbr.ParseRange(%d, %d): error: %s", trng.start, trng.end, err)
			}

			got, err := ioutil.ReadAll(rbr)
			if err != nil {
				t.Errorf("ioutil.ReadAll(pres.Body): got error: %s", err)
			}

			want := bwnt[trng.start : trng.end+1]
			if trng.end == len(bwnt) {
				want = bwnt[trng.start:]
			}

			if len(got) != len(want) {
				t.Errorf("len: got: %d, want %d", len(got), len(want))
			}

			if !bytes.Equal(got, want) {
				t.Errorf("bytes.Equal for b[%d:%d]:\ngot:\n%s\nwant:\n%s", trng.start, trng.end+1, got, want)
			}

		})
	}
}
