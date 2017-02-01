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

// Package archive provides functionality around the creation and management
// of archives of HTTP responses.
package archive

import (
	"net/http"
)

type Archive struct {
    
}

func New(path string) *Archive {
    // stat the path to see if the archive exists

}

func (a *Archive) initialize() error {

}

func (a *Archive) Open() {
    // unzip to tmp
    // archive/
    //        /bolt.db
    //        / buckets - {chunked, unchunked}
    //        /files/

}

func (a *Archive) Close() string
   // zip it up

}

func (a *Archive) Append(key string, res *http.Response) {

}

func (a *Archive) Query(key string) *http.Response{
    
}
