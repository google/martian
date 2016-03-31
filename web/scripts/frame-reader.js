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

function FrameReader(view) {
  this._frames = {
    1: 'header-frame',
    2: 'data-frame',
  };

  this._scopes = {
    1: 'request',
    2: 'response',
  }

  this.n = 0;
  this.view = view;
}

FrameReader.prototype.read = function() {
  var parser = new Parser(this.view, this.n);
  var frame = {};

  frame.magic = parser.uint32();
  frame.type = this._frames[parser.uint8()];
  frame.scope = this._scopes[parser.uint8()];
  frame.id = parser.string(8);

  switch (frame.type) {
    case 'header-frame':
      var nameLength = parser.uint32();
      var valueLength = parser.uint32();

      console.log('header-frame length', nameLength, valueLength);

      frame.name = parser.string(nameLength);
      frame.value = parser.string(valueLength);

      break;
      case 'data-frame':
        frame.index = parser.uint32();
        frame.end = parser.uint8();

        var dataLength = parser.uint32();
        console.log('data-frame length', dataLength);
        frame.data = parser.bytes(dataLength);

        break;
      default:
        console.log('received unknown frame: ' + frame.id);
  };

  this.n = parser._offset;

  return frame;
};
