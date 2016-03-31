'use strict'
var arrayify = require('array-back')
var option = require('./option')
var findReplace = require('find-replace')

/**
 * Handles parsing different argv notations
 *
 * @module argv
 * @private
 */

class Argv extends Array {
  constructor (argv) {
    super()

    if (argv) {
      argv = arrayify(argv)
    } else {
      /* if no argv supplied, assume we are parsing process.argv */
      argv = process.argv.slice(0)
      argv.splice(0, 2)
    }

    this.load(argv)
  }

  load (array) {
    arrayify(array).forEach(item => this.push(item))
  }

  clear () {
    this.length = 0
  }

  /* expand --option=name style args */
  expandOptionEqualsNotation () {
    var optEquals = option.optEquals
    if (this.some(optEquals.test.bind(optEquals))) {
      var expandedArgs = []
      this.forEach(arg => {
        var matches = arg.match(optEquals.re)
        if (matches) {
          expandedArgs.push(matches[1], matches[2])
        } else {
          expandedArgs.push(arg)
        }
      })
      this.clear()
      this.load(expandedArgs)
    }
  }

  /* expand getopt-style combined options */
  expandGetoptNotation () {
    var combinedArg = option.combined
    var hasGetopt = this.some(combinedArg.test.bind(combinedArg))
    if (hasGetopt) {
      findReplace(this, combinedArg.re, arg => {
        arg = arg.slice(1)
        return arg.split('').map(letter => '-' + letter)
      })
    }
  }

  validate (definitions) {
    var invalidOption

    var optionWithoutDefinition = this
      .filter(arg => option.isOption(arg))
      .some(arg => {
        if (definitions.get(arg) === undefined) {
          invalidOption = arg
          return true
        }
      })
    if (optionWithoutDefinition) {
      halt(
        'UNKNOWN_OPTION',
        'Unknown option: ' + invalidOption
      )
    }
  }
}

function halt (name, message) {
  var err = new Error(message)
  err.name = name
  throw err
}

module.exports = Argv
