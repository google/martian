var test = require('tape')
var cliArgs = require('../')

test('err-invalid-definition: throws when no definition.name specified', function (t) {
  var optionDefinitions = [
    { something: 'one' },
    { something: 'two' }
  ]
  var argv = [ '--one', '--two' ]
  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'NAME_MISSING')
  }
  t.end()
})

test('err-invalid-definition: throws if dev set a numeric alias', function (t) {
  var optionDefinitions = [
    { name: 'colours', alias: '1' }
  ]
  var argv = [ '--colours', 'red' ]

  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'INVALID_ALIAS')
  }

  t.end()
})

test('err-invalid-definition: throws if dev set an alias of "-"', function (t) {
  var optionDefinitions = [
    { name: 'colours', alias: '-' }
  ]
  var argv = [ '--colours', 'red' ]

  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'INVALID_ALIAS')
  }

  t.end()
})

test('err-invalid-definition: multi-character alias', function (t) {
  var optionDefinitions = [
    { name: 'one', alias: 'aa' }
  ]
  var argv = [ '--one', 'red' ]

  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'INVALID_ALIAS')
  }

  t.end()
})

test('err-invalid-definition: invalid type values', function (t) {
  var argv = [ '--one', 'something' ]
  try {
    cliArgs([ { name: 'one', type: 'string' } ]).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'INVALID_TYPE')
  }

  try {
    cliArgs([ { name: 'one', type: 234 } ]).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'INVALID_TYPE')
  }

  try {
    cliArgs([ { name: 'one', type: {} } ]).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'INVALID_TYPE')
  }

  t.doesNotThrow(function () {
    cliArgs([ { name: 'one', type: function () {} } ]).parse(argv)
  }, /invalid/i)

  t.end()
})

test('err-invalid-definition: value without option definition', function (t) {
  var optionDefinitions = [
    { name: 'one', type: Number }
  ]

  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--one', '1' ]),
    { one: 1 }
  )

  try {
    cliArgs(optionDefinitions).parse([ '--one', '--two' ])
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'UNKNOWN_OPTION')
  }

  try {
    cliArgs(optionDefinitions).parse([ '--one', '2', '--two', 'two' ])
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'UNKNOWN_OPTION')
  }

  try {
    cliArgs(optionDefinitions).parse([ '-a', '2' ])
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'UNKNOWN_OPTION')
  }

  try {
    cliArgs(optionDefinitions).parse([ '-sdf' ])
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'UNKNOWN_OPTION', 'getOpts')
  }

  t.end()
})

test('err-invalid-definition: duplicate name', function (t) {
  var optionDefinitions = [
    { name: 'colours' },
    { name: 'colours' }
  ]
  var argv = [ '--colours', 'red' ]

  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'DUPLICATE_NAME')
  }

  t.end()
})

test('err-invalid-definition: duplicate alias', function (t) {
  var optionDefinitions = [
    { name: 'one', alias: 'a' },
    { name: 'two', alias: 'a' }
  ]
  var argv = [ '--one', 'red' ]

  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'DUPLICATE_ALIAS')
  }

  t.end()
})

test('err-invalid-definition: multiple defaultOption', function (t) {
  var optionDefinitions = [
    { name: 'one', defaultOption: true },
    { name: 'two', defaultOption: true }
  ]
  var argv = [ '--one', 'red' ]

  try {
    cliArgs(optionDefinitions).parse(argv)
    t.fail()
  } catch (err) {
    t.strictEqual(err.name, 'DUPLICATE_DEFAULT_OPTION')
  }

  t.end()
})
