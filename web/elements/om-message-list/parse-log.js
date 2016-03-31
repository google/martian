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

