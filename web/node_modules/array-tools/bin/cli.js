#!/usr/bin/env node
'use strict'
var a = require('../')
var domain = require('domain')
var ansi = require('ansi-escape-sequences')
var util = require('util')
var collectJson = require('collect-json')

if (process.argv.length < 3) {
  console.error('Usage:')
  console.error('$ cat <json array> | array-tools <method> <args...>')
  process.exit(1)
}

process.argv.splice(0, 2)
var method = process.argv.shift()
var args = process.argv.slice(0)
args = args.map(function (arg) {
  return arg.replace(/\\n/g, '\n')
})

switch (method) {
  /* convert map arg to a function */
  case 'map':
    var funcBody = args.shift()
    var mapFunction = eval(util.format('(function mapFunction(item){%s})', funcBody))
    args.unshift(mapFunction)
    break
}

function processInput (input) {
  var arr = a(input)
  var result

  switch (method) {
    case 'pick':
      result = arr[method](args)
      break
    default:
      result = arr[method].apply(arr, args)
  }

  if (result._data) result = result.val()

  /* certain methods don't output JSON */
  if (a.contains([ 'join' ], method)) {
    return result + '\n'
  } else {
    return JSON.stringify(result, null, '  ') + '\n'
  }
}

var d = domain.create()
d.on('error', function (err) {
  if (err.code === 'EPIPE') return // don't care
  console.error(ansi.format('Error: ' + err.stack, 'red'))
})
d.run(function () {
  process.stdin
    .pipe(collectJson(processInput))
    .pipe(process.stdout)
})
