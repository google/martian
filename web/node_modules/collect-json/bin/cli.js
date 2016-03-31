#!/usr/bin/env node
"use strict";
var collectJson = require("../");
var ansi = require("ansi-escape-sequences");

var funcBody = process.argv[2];
if (!funcBody){
    halt("Usage:\n$ cat <json> | collect-json <through function body>");
}
var throughFunction = new Function("json", funcBody);

process.stdin
    .pipe(collectJson(throughFunction))
    .on("error", halt)
    .pipe(process.stdout);

function halt(msg){
    if (msg instanceof Error){
        msg = msg.message
    }
    console.error(ansi.format(msg, "red"));
    process.exit(1);
}