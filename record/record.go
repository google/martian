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

// Package record writes to a file that contains a hash of requests to
// responses.
package record

import (
	"bytes"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/google/martian/marbl"
)

// Recorder is a response modifier that writes responses to a boltdb.
type Recorder struct {
	db     string
	keygen func(*http.Request) []byte
}

// NewRecorder returns a Recorder with an keygen function initialized to
// use the request path as the response key.
func NewRecorder(dbpath string) *Recorder {
	kg := func(req *http.Request) []byte {
		return []byte(req.URL.String())
	}
	return &Recorder{
		keygen: kg,
		db:     dbpath,
	}
}

// SetKeyGenFunc sets the keygen function used to generate keys for responses.
func (r *Recorder) SetKeyGenFunc(keygen func(*http.Request) []byte) {
	r.keygen = keygen
}

// ModifyResponse executes m.keygen and saves the response keyed to the
// key generated.
func (r *Recorder) ModifyResponse(res *http.Response) error {
	db, err := bolt.Open(r.db, 0644, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if res.Header.Get("Content-Length") == "-1" {
		// chunked response

	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("responses"))
		if err != nil {
			return err
		}

		key := r.keygen(res.Request)

		var b bytes.Buffer
		m := marbl.NewStream(b)

		err = bucket.Put(key, value)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}
