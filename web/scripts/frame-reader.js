
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

function Parser(view, offset) {
  this._view = view;
  this._offset = offset || 0;
}

Parser.prototype.uint8 = function() {
    var value = this._view.getUint8(this._offset);
    this._offset++;

    return value;
}

Parser.prototype.uint32 = function() {
    var value = this._view.getUint32(this._offset);
    this._offset += 4;

    return value;
}

Parser.prototype.string = function(length) {
    var bytes = this.bytes(length);
    var decoder = new TextDecoder("utf-8");

    return decoder.decode(bytes);
}

Parser.prototype.bytes = function(length) {
    var data = this._view.buffer.slice(this._offset, this._offset+length);
    this._offset += length;

    return new Uint8Array(data);
}

function FrameReader() {
  this._frames = {
    1: 'header-frame',
    2: 'data-frame',
  };

  this._scopes = {
    1: 'request',
    2: 'response',
  }

  this.n = 0;
}

FrameReader.prototype.read = function(arraybuffer_view, offset) {
  var view = new DataView(arraybuffer_view);

  var parser = new Parser(view, offset);
  var frame = {};

  frame.type = this._frames[parser.uint8()];
  frame.scope = this._scopes[parser.uint8()];
  frame.id = parser.string(8);

  switch (frame.type) {
    case 'header-frame':
      var nameLength = parser.uint32();
      var valueLength = parser.uint32();


      frame.name = parser.string(nameLength);
      frame.value = parser.string(valueLength);

      break;
    case 'data-frame':
      frame.index = parser.uint32();
      frame.end = parser.uint8();

      var dataLength = parser.uint32();
      frame.data = parser.bytes(dataLength);

      break;
    default:
  };

  this.n = parser._offset;

  return frame;
};
