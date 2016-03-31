'use strict'
var test = require('tape')
var columnLayout = require('../')

test('columnLayout.lines(data, options): no data', function (t) {
  t.deepEqual(columnLayout.lines([]), [])
  t.deepEqual(columnLayout.lines(), [])
  t.end()
})
