'use strict'
var test = require('tape')
var columnLayout = require('../')

test('columnLayout(data, options)', function (t) {
  var fixture = require('./fixture/simple-viewWidth')
  var result =
'<row 1 column one .. ><r1 c2           >\n\
<.. ..               ><                >\n\
<r2 c1               ><row two column 2>\n\
'
  t.strictEqual(columnLayout(fixture.data, fixture.options), result)
  t.end()
})

test('columnLayout.lines(data, options)', function (t) {
  var fixture = require('./fixture/simple-viewWidth')
  var result = [
    "<row 1 column one .. ><r1 c2           >",
    "<.. ..               ><                >",
    "<r2 c1               ><row two column 2>"
  ]

  t.deepEqual(columnLayout.lines(fixture.data, fixture.options), result)
  t.end()
})

test('columnLayout.lines(data, options)', function (t) {
  var fixture = require('./fixture/primatives')
  var result = [
    "<row 1 column one .. .. ..><3000>",
    "<true                     ><null>",
    "<[object Object]          ><    >"
  ]

  t.deepEqual(columnLayout.lines(fixture.data, fixture.options), result)
  t.end()
})
