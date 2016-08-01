
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

// Package martian provides an HTTP/1.1 proxy with an API for configurable
// request and response modifiers.

importScripts('/scripts/frame-reader.js');

self.addEventListener('message', function(e) {
  var view = new DataView(e.data);
  var reader = new FrameReader(view);

  while (reader.n < view.byteLength) {
    var frame = reader.read();
    console.log('read frame', frame, reader.n, view.byteLength);
    self.postMessage(frame);
  }

  self.close();
});

