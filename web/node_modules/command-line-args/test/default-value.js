var test = require('tape')
var cliArgs = require('../')

test('default value', function (t) {
  t.deepEqual(cliArgs([ { name: 'one' }, { name: 'two', defaultValue: 'two' } ]).parse([ '--one', '1' ]), {
    one: '1',
    two: 'two'
  })
  t.deepEqual(cliArgs([{ name: 'two', defaultValue: 'two' }]).parse([]), {
    two: 'two'
  })
  t.deepEqual(cliArgs([{ name: 'two', defaultValue: 'two' }]).parse([ '--two', 'zwei' ]), {
    two: 'zwei'
  })
  t.deepEqual(
    cliArgs([{ name: 'two', multiple: true, defaultValue: ['two', 'zwei'] }]).parse([ '--two', 'duo' ]),
    { two: [ 'duo' ] }
  )

  t.end()
})

test('default value', function (t) {
  var defs = [{ name: 'two', multiple: true, defaultValue: ['two', 'zwei'] }]
  var result = cliArgs(defs).parse([])
  t.deepEqual(result, { two: [ 'two', 'zwei' ] })
  t.end()
})

test('default value: array as defaultOption', function (t) {
  var defs = [
    { name: 'two', multiple: true, defaultValue: ['two', 'zwei'], defaultOption: true }
  ]
  var argv = [ 'duo' ]
  t.deepEqual(cliArgs(defs).parse(argv), { two: [ 'duo' ] })
  t.end()
})

test('default value: falsy default values', function (t) {
  var defs = [
    { name: 'one', defaultValue: 0 },
    { name: 'two', defaultValue: false }
  ]

  var argv = []
  t.deepEqual(cliArgs(defs).parse(argv), {
    one: 0,
    two: false
  })
  t.end()
})
