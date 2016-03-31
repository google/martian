var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'verbose', alias: 'v' },
  { name: 'colour', alias: 'c' },
  { name: 'number', alias: 'n' },
  { name: 'dry-run', alias: 'd' }
]

test('instatiate: with new', function (t) {
  var argv = [ '-v' ]
  var cli = new cliArgs(optionDefinitions) // eslint-disable-line new-cap
  t.deepEqual(cli.parse(argv), {
    verbose: true
  })
  t.end()
})

test('instatiate: without new', function (t) {
  var argv = [ '-v' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    verbose: true
  })
  t.end()
})
