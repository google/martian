/**
 * @fileoverview Description of this file.
 */
var assert = require('assert');
var FrameReader = require('../scripts/frame-reader.js').FrameReader;

suite('frame-reader', function() {
  var reader;

  // Frame types.
  var HEADER = 0x1;
  var DATA = 0x2;
  // Message types.
  var REQUEST = 0x1;
  var RESPONSE = 0x2;

  var newHeaderFrame = function(headerName, headerValue) {
    var nameLen = headerName.length;
    var valueLen = headerValue.length;
    var bufferSize = 8 + 8 + 8 + 32 + 32 + nameLen + valueLen;
    var buffer = new ArrayBuffer(bufferSize);
    var frame = new DataView(buffer);

    frame.setUint8(0, HEADER);
    frame.setUint8(1, REQUEST);
    frame.setUint8(2, 'TESTID01');
    frame.setUint32(3, nameLen);
    frame.setUint32(7, valueLen);
    var offset = 8;
    for (var c in headerName) {
      frame.setUint8(offset, c);
      offset++;
    }
    for (var c in headerValue) {
      frame.setUint8(offset, c);
      offset++;
    }
    return buffer;
  };

  setup(function() {
    reader = new FrameReader();
    //var td = new TextDecoder("utf-8");
  });

  test('a', function() {
    var frame = newHeaderFrame('someheader', 'someheadervalue');
    reader.read(frame);
  });
  test('b', function() {
    var frame = newHeaderFrame('someheader', 'someheadervalue');
    reader.read(frame);
  });
});
