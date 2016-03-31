var test = require('tape')
var cliArgs = require('../')

test('defaultOption: string', function (t) {
  var optionDefinitions = [
    { name: 'files', defaultOption: true }
  ]
  var argv = [ 'file1', 'file2' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    files: 'file2'
  })
  t.end()
})

test('defaultOption: multiple string', function (t) {
  var optionDefinitions = [
    { name: 'files', defaultOption: true, multiple: true }
  ]
  var argv = [ 'file1', 'file2' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    files: [ 'file1', 'file2' ]
  })
  t.end()
})

test('defaultOption: after a boolean', function (t) {
  var cli = cliArgs([
    { name: 'one', type: Boolean },
    { name: 'two', defaultOption: true }
  ])
  t.deepEqual(
    cli.parse([ '--one', 'sfsgf' ]),
    { one: true, two: 'sfsgf' }
  )

  t.end()
})

test('defaultOption: multiple defaultOption values between other arg/value pairs', function (t) {
  var optionDefinitions = [
    { name: 'one' },
    { name: 'two' },
    { name: 'files', defaultOption: true, multiple: true }
  ]
  var argv = [ '--one', '1', 'file1', 'file2', '--two', '2' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    one: '1',
    two: '2',
    files: [ 'file1', 'file2' ]
  })
  t.end()
})

test('defaultOption: multiple defaultOption values between other arg/value pairs 2', function (t) {
  var optionDefinitions = [
    { name: 'one', type: Boolean },
    { name: 'two' },
    { name: 'files', defaultOption: true, multiple: true }
  ]
  var argv = [ 'file0', '--one', 'file1', '--files', 'file2', '--two', '2', 'file3' ]
  t.deepEqual(cliArgs(optionDefinitions).parse(argv), {
    one: true,
    two: '2',
    files: [ 'file0', 'file1', 'file2', 'file3' ]
  })
  t.end()
})

test('defaultOption: floating args present but no defaultOption', function (t) {
  var cli = cliArgs([
    { name: 'one', type: Boolean }
  ])
  t.deepEqual(
    cli.parse([ 'aaa', '--one', 'aaa', 'aaa' ]),
    { one: true }
  )

  t.end()
})
