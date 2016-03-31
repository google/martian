'use strict'
var arrayify = require('array-back')
var option = require('./option')
var Definition = require('./definition')
var t = require('typical')

/**
 * @class
 * @alias module:definitions
 */
class Definitions extends Array {
  constructor (definitions) {
    super()
    arrayify(definitions).forEach(def => this.push(new Definition(def)))
    this.validate()
  }

  /**
   * validate option definitions
   * @returns {string}
   */
  validate (argv) {
    var someHaveNoName = this.some(def => !def.name)
    if (someHaveNoName) {
      halt(
        'NAME_MISSING',
        'Invalid option definitions: the `name` property is required on each definition'
      )
    }

    var someDontHaveFunctionType = this.some(def => def.type && typeof def.type !== 'function')
    if (someDontHaveFunctionType) {
      halt(
        'INVALID_TYPE',
        'Invalid option definitions: the `type` property must be a setter fuction (default: `Boolean`)'
      )
    }

    var invalidOption

    var numericAlias = this.some(def => {
      invalidOption = def
      return t.isDefined(def.alias) && t.isNumber(def.alias)
    })
    if (numericAlias) {
      halt(
        'INVALID_ALIAS',
        'Invalid option definition: to avoid ambiguity an alias cannot be numeric [--' + invalidOption.name + ' alias is -' + invalidOption.alias + ']'
      )
    }

    var multiCharacterAlias = this.some(def => {
      invalidOption = def
      return t.isDefined(def.alias) && def.alias.length !== 1
    })
    if (multiCharacterAlias) {
      halt(
        'INVALID_ALIAS',
        'Invalid option definition: an alias must be a single character'
      )
    }

    var hypenAlias = this.some(def => {
      invalidOption = def
      return def.alias === '-'
    })
    if (hypenAlias) {
      halt(
        'INVALID_ALIAS',
        'Invalid option definition: an alias cannot be "-"'
      )
    }

    var duplicateName = hasDuplicates(this.map(def => def.name))
    if (duplicateName) {
      halt(
        'DUPLICATE_NAME',
        'Two or more option definitions have the same name'
      )
    }

    var duplicateAlias = hasDuplicates(this.map(def => def.alias))
    if (duplicateAlias) {
      halt(
        'DUPLICATE_ALIAS',
        'Two or more option definitions have the same alias'
      )
    }

    var duplicateDefaultOption = hasDuplicates(this.map(def => def.defaultOption))
    if (duplicateDefaultOption) {
      halt(
        'DUPLICATE_DEFAULT_OPTION',
        'Only one option definition can be the defaultOption'
      )
    }
  }

  createOutput () {
    var output = {}
    this.forEach(def => {
      if (t.isDefined(def.defaultValue)) output[def.name] = def.defaultValue
      if (Array.isArray(output[def.name])) {
        output[def.name]._initial = true
      }
    })
    return output
  }

  get (arg) {
    return option.short.test(arg)
      ? this.find(def => def.alias === option.short.name(arg))
      : this.find(def => def.name === option.long.name(arg))
  }

  getDefault () {
    return this.find(def => def.defaultOption === true)
  }

  isGrouped () {
    return this.some(def => def.group)
  }

  whereGrouped () {
    return this.filter(containsValidGroup)
  }
  whereNotGrouped () {
    return this.filter(def => !containsValidGroup(def))
  }

}

function halt (name, message) {
  var err = new Error(message)
  err.name = name
  throw err
}

function containsValidGroup (def) {
  return arrayify(def.group).some(group => group)
}

function hasDuplicates (array) {
  var items = {}
  for (var i = 0; i < array.length; i++) {
    var value = array[i]
    if (items[value]) {
      return true
    } else {
      if (t.isDefined(value)) items[value] = true
    }
  }
}

/**
@module definitions
@private
*/
module.exports = Definitions
