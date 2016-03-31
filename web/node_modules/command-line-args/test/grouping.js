var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'one', group: 'a' },
  { name: 'two', group: 'a' },
  { name: 'three', group: 'b' }
]

test('groups', function (t) {
  var cli = cliArgs(optionDefinitions)
  t.deepEqual(cli.parse([ '--one', '1', '--two', '2', '--three', '3' ]), {
    a: {
      one: '1',
      two: '2'
    },
    b: {
      three: '3'
    },
    _all: {
      one: '1',
      two: '2',
      three: '3'
    }
  })

  t.end()
})

test('groups: multiple and _none', function (t) {
  var optionDefinitions = [
    { name: 'one', group: ['a', 'f'] },
    { name: 'two', group: ['a', 'g'] },
    { name: 'three' }
  ]

  var cli = cliArgs(optionDefinitions)
  t.deepEqual(cli.parse([ '--one', '1', '--two', '2', '--three', '3' ]), {
    a: {
      one: '1',
      two: '2'
    },
    f: {
      one: '1'
    },
    g: {
      two: '2'
    },
    _none: {
      three: '3'
    },
    _all: {
      one: '1',
      two: '2',
      three: '3'
    }
  })

  t.end()
})
