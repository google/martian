'use strict'
var test = require('tape')
var columnLayout = require('..')

test('columnLayout.table()', function(t){
  var fixture = require('./fixture/simple-viewWidth')
  var table = columnLayout.table(fixture.data, fixture.options)

  t.strictEqual(table.rows.length, 2)
  t.strictEqual(table.columns.length, 2)
  t.end()
})

test('table.getWrapped()', function(t){
  var fixture = require('./fixture/simple-viewWidth')
  var table = columnLayout.table(fixture.data, fixture.options)

  t.deepEqual(table.getWrapped(), [
    [ ['row 1 column one ..', '.. ..'], ['r1 c2'] ],
    [ ['r2 c1'], ['row two column 2'] ]
  ])
  t.end()
})

test('table.getLines()', function(t){
  var fixture = require('./fixture/simple-viewWidth')
  var table = columnLayout.table(fixture.data, fixture.options)

  t.deepEqual(table.getLines(), [
    [ 'row 1 column one ..', 'r1 c2' ],
    [ '.. ..', '' ],
    [ 'r2 c1', 'row two column 2' ]
  ])
  t.end()
})

test('table.renderLines()', function(t){
  var fixture = require('./fixture/simple-viewWidth')
  var table = columnLayout.table(fixture.data, fixture.options)

  t.deepEqual(table.renderLines(), [
    '<row 1 column one .. ><r1 c2           >',
    '<.. ..               ><                >',
    '<r2 c1               ><row two column 2>'
  ])
  t.end()
})
