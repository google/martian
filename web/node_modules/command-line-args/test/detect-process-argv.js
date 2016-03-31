var test = require('tape')
var cliArgs = require('../')

test('detect process.argv: should automatically remove first two argv items', function (t) {
  process.argv = [ 'node', 'filename', '--one', 'eins' ]
  t.deepEqual(cliArgs({ name: 'one' }).parse(process.argv), {
    one: 'eins'
  })
  t.end()
})

test('process.argv is left untouched', function (t) {
  process.argv = [ 'node', 'filename', '--one', 'eins' ]
  t.deepEqual(cliArgs({ name: 'one' }).parse(), {
    one: 'eins'
  })
  t.deepEqual(process.argv, [ 'node', 'filename', '--one', 'eins' ])
  t.end()
})
