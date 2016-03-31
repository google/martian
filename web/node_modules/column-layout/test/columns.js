'use strict'
var test = require('tape')
var Columns = require('../').Columns

test('columns.autoSize(contentColumns, viewWidth)', function (t) {
  var columns = new Columns([
    { name: 'one', contentWidth: 10, contentWrappable: true },
    { name: 'two', contentWidth: 20, contentWrappable: true }
  ])

  columns.viewWidth = 30
  columns.autoSize()
  t.strictEqual(columns[0].generatedWidth, 12)
  t.strictEqual(columns[1].generatedWidth, 18)

  t.end()
})
