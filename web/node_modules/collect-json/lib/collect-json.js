"use strict";
var t = require("typical");
var connect = require("stream-connect");
var through = require("stream-via");
var collect = require("collect-all");

/**
Returns a stream which becomes readable with a single value once all (valid) JSON is received.

## Synopsis
At the terminal
```sh
$ echo '"simple"' | collect-json "return json + ' example\n'"
simple example
```
@module collect-json
*/
module.exports = collectJson;
collectJson.async = collectJsonAsync;

/**
@param [throughFunction] {function} - an optional function to transform the data before passing it on.
@return {external:Duplex}
@alias module:collect-json
@example 
An example command-line client script - JSON received at stdin is stamped with `received` then written to stdout. 
```js
var collectJson = require("collect-json");

process.stdin
    .pipe(collectJson(function(json){
        json.received = true;
        return JSON.stringify(json);
    }))
    .on("error", function(err){
        // input from stdin failed to parse
    })
    .pipe(process.stdout);
```
*/
function collectJson(throughFunction){
    var stream = collect({
        through: function(data){
            try {
                var json = JSON.parse(data);
            } catch(err){
                err.input = data;
                err.message = "Error parsing input JSON: " + err.message;
                throw err;
            }
            return json;
        },
        objectMode: true
    });
    if (throughFunction){
        return connect(stream, through(throughFunction, { objectMode: true }));
    } else {
        return stream;
    }
}

function collectJsonAsync(throughFunction, options){
}

/**
@external Duplex
@see https://nodejs.org/api/stream.html#stream_class_stream_duplex
*/
