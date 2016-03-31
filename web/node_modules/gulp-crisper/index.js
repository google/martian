'use strict';
var path = require('path');
var gutil = require('gulp-util');
var through = require('through2');
var crisper = require('crisper');
var oassign = require('object-assign');

function splitFile(file, filename, contents) {
  return new gutil.File({
    cwd: file.cwd,
    base: file.base,
    path: path.join(path.dirname(file.path), filename),
    contents: new Buffer(contents)
  });
}

function getFilename(filepath) {
  var basename = path.basename(filepath, path.extname(filepath));
  return {
    html: basename + '.html',
    js: basename + '.js'
  };
}

module.exports = function (opts) {
  return through.obj(function (file, enc, cb) {
    if (file.isNull()) {
      cb(null, file);
      return;
    }

    if (file.isStream()) {
      cb(new gutil.PluginError('gulp-crisper', 'Streaming not supported'));
      return;
    }

    var splitfile = getFilename(file.path);
    var split = crisper(oassign({}, opts, {
      source: file.contents.toString(),
      jsFileName: splitfile.js
    }));
    var stream = this;

    Object.keys(split).forEach(function(type) {
      if (split[type]) {
        stream.push(splitFile(file, splitfile[type], split[type]));
      }
    });

    cb();
  });
};
