"use strict";
var Transform = require("stream").Transform;
var util = require("util");
var t = require("typical");
var connect = require("stream-connect");
var through = require("stream-via");

/**
Returns a stream which becomes readable once all input is received

@module collect-all
*/
module.exports = collect;

function CollectTransform(options){
    if (!(this instanceof CollectTransform)) return new CollectTransform(options);
    Transform.call(this, options);
    this.buf = new Buffer(0);
}
util.inherits(CollectTransform, Transform);

CollectTransform.prototype._transform = function(chunk, enc, done){
    if (chunk){
        this.buf = Buffer.concat([ this.buf, chunk ]);
    }
    done();
};

CollectTransform.prototype._flush = function(){
    this.push(this.buf);
    this.push(null);
};

/**
@param [options] {object}
@param [options.through] {function}
@return {external:Duplex}
@alias module:collect-all
*/
function collect(options){
    options = options || {};
    if (!t.isPlainObject(options)) throw Error("Invalid options");
    if (options.through){
        return connect(CollectTransform(), through(options.through, options));
    } else {
        return new CollectTransform();
    }
}

/**
@external Duplex
@see https://nodejs.org/api/stream.html#stream_class_stream_duplex
*/
