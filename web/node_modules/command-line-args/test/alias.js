var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'verbose', alias: 'v' },
  { name: 'colour', alias: 'c' },
  { name: 'number', alias: 'n' },
  { name: 'dry-run', alias: 'd' }
]

test('alias: one boolean', function (t) {
  var argv = [ '-v' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    verbose: true
  })
  t.end()
})

test('alias: one --this-type boolean', function (t) {
  var argv = [ '-d' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    'dry-run': true
  })
  t.end()
})

test('alias: one boolean, one string', function (t) {
  var argv = [ '-v', '-c' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    verbose: true,
    colour: true
  })
  t.end()
})
