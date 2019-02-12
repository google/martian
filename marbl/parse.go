package marbl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/martian/log"
)

type Parse struct {
	reader *Reader
}

// New provides a *Cache with an initialized marbl.Reader.
func New(r io.Reader) *Parse {
	mr := NewReader(r)
	return &Parse{reader: mr}
}

// Responses scans the Cache.reader and reconstructs the responses as a
// []*http.Response. Requests are reconstructed and are associated with their
// responses in Response.Request. Any requests that do not have a corresponding
// response are ignored.
func (p *Parse) Responses() ([]*http.Response, error) {
	respids := make(map[string]*http.Response, 0)
	bbufids := make(map[string]*bytes.Buffer, 0)

	for {
		frame, err := p.reader.ReadFrame()
		if frame == nil && err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("c.reader.ReadFrame(): %s", err)
		}

		if frame.FrameType() == DataFrame {
			dfr, ok := frame.(Data)
			if !ok {
				return nil, fmt.Errorf("frame.(marbl.Data):  %s", err)
			}

			resp, ok := respids[dfr.ID]
			if !ok {
				resp = &http.Response{Request: &http.Request{}}
				respids[dfr.ID] = resp
			}

			bbuf, ok := bbufids[dfr.ID]
			if !ok {
				bbuf = new(bytes.Buffer)
				bbufids[dfr.ID] = bbuf
			}

			switch dfr.MessageType {
			case Request:
			case Response:
				bbuf.Write(dfr.Data)

				continue
			default:
				log.Infof("parse.Responses: Could not determine marbl.MessageType for DataFrame ID: %s", dfr.ID)
			}

			continue
		}

		hdr, ok := frame.(Header)
		if !ok {
			return nil, fmt.Errorf("frame.(marbl.Header): %s", err)
		}

		resp, ok := respids[hdr.ID]
		if !ok {
			resp = &http.Response{Request: &http.Request{URL: &url.URL{}}}
			resp.Header = make(map[string][]string, 0)
			resp.Request.Header = make(map[string][]string, 0)
			respids[hdr.ID] = resp
		}

		switch hdr.MessageType {
		case Request:
			switch hdr.Name {
			case ":method":
				resp.Request.Method = hdr.Value
			case ":scheme":
				resp.Request.URL.Scheme = hdr.Value
			case ":authority":
				resp.Request.Host = hdr.Value
				resp.Request.URL.Host = hdr.Value
			case ":path":
				resp.Request.URL.Path = hdr.Value
			case ":query":
				resp.Request.URL.RawQuery = hdr.Value
			case ":fragment":
				//TODO(bramha): add fragment support to marbl.LogRequest
			default:
				if strings.HasPrefix(hdr.Name, ":") {

					continue
				}
				resp.Request.Header.Set(hdr.Name, hdr.Value)
			}

			continue
		case Response:
			if strings.HasPrefix(hdr.Name, ":") {

				continue
			}
			resp.Header.Set(hdr.Name, hdr.Value)

			continue

		default:
			log.Infof("marbl.Parse.Responses: Could not determine marbl.MessageType for HeaderFrame ID: %s", hdr.ID)
		}
	}

	resps := make([]*http.Response, 0, len(respids))
	for id, res := range respids {
		if body, ok := bbufids[id]; ok {
			log.Infof("marbl.Parse.Responses: body.Len(): %d for res.Request: %s", body.Len(), res.Request.URL.String())
			res.ContentLength = int64(body.Len())
			res.Body = ioutil.NopCloser(body)
		}

		resps = append(resps, res)
	}

	return resps, nil
}

// RequestURLs scans and returns all for all request URLs. The order
// of the URLs is arbitrary and does not reflect the order in which the
// requests occured.
func (p *Parse) RequestURLs() ([]*url.URL, error) {
	reqs := make(map[string]*url.URL)

	for {
		frame, err := p.reader.ReadFrame()
		if frame == nil && err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("marbl.Parse.ReadRequestURLs: %s", err)
		}

		if frame.FrameType() != HeaderFrame {
			continue
		}

		hdr, ok := frame.(Header)
		if !ok {
			return nil, fmt.Errorf("marbl.Parse.ReadRequestURLs: %s", err)
		}

		if hdr.MessageType != Request {
			continue
		}

		if _, ok := reqs[hdr.ID]; !ok {
			reqs[hdr.ID] = &url.URL{}
		}

		req := reqs[hdr.ID]

		// Headers that do not start with a colon are normal HTTP traffic headers
		// instead of meta-data headers, and should be ignored.
		if !strings.HasPrefix(hdr.Name, ":") {
			continue
		}

		switch hdr.Name {
		case ":method":
			// TODO(bramha): determine whether to only support GET request URLs.
			continue
		case ":scheme":
			req.Scheme = hdr.Value
			continue
		case ":authority":
			req.Host = hdr.Value
			continue
		case ":path":
			req.Path = hdr.Value
			continue
		case ":query":
			req.RawQuery = hdr.Value
			continue
		case ":fragment":
			//TODO(bramha): add fragment support to marbl.LogRequest
		case ":proto":
		case ":remote":
		case ":timestamp":
			continue
		default:
			log.Infof("marbl.Parse.ReadRequestURLs: unknown header Name key: %s", hdr.Name)
		}
	}

	urls := make([]*url.URL, 0, len(reqs))
	for _, req := range reqs {
		urls = append(urls, req)
	}

	return urls, nil
}
