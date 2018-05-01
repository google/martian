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
package server

import (
	"net/http"
	"testing"

	"github.com/google/martian"

	"github.com/google/martian/fifo"
	"github.com/google/martian/har"
)

func TestServer(t *testing.T) {
	t.Skip("not a real test")

	phar := har.NewLogger()

	server, _ := NewServer("somenamed", 8080, 8888,
		EnableTrafficShaping(),
		AllowCORS(),
		EnableMITM("cert", "key"),
		SetPremodificationLogger(har.NewLogger(), map[string]func(martian.Logger) http.HandlerFunc{
			"/logs": har.ExportHandlerFunc,
		}),
		AddModifiers(fifo.NewGroup(), "/config", "/config/reset"),
	)

	server.Start()
}
