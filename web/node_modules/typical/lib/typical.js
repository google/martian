'use strict'

/**
For type-checking Javascript values.
@module typical
@typicalname t
@example
var t = require("typical")
*/
exports.isNumber = isNumber
exports.isString = isString
exports.isBoolean = isBoolean
exports.isPlainObject = isPlainObject
exports.isArrayLike = isArrayLike
exports.isObject = isObject
exports.isDefined = isDefined
exports.isFunction = isFunction

/**
Returns true if input is a number
@param {*} - the input to test
@returns {boolean}
@static
@example
> t.isNumber(0)
true
> t.isNumber(1)
true
> t.isNumber(1.1)
true
> t.isNumber(0xff)
true
> t.isNumber(0644)
true
> t.isNumber(6.2e5)
true
> t.isNumber(NaN)
false
> t.isNumber(Infinity)
false
*/
function isNumber (n) {
  return !isNaN(parseFloat(n)) && isFinite(n)
}

/**
A plain object is a simple object literal, it is not an instance of a class. Returns true if the input `typeof` is `object` and directly decends from `Object`.

@param {*} - the input to test
@returns {boolean}
@static
@example
> t.isPlainObject({ clive: "hater" })
true
> t.isPlainObject(new Date())
false
> t.isPlainObject([ 0, 1 ])
false
> t.isPlainObject(1)
false
> t.isPlainObject(/test/)
false
*/
function isPlainObject (input) {
  return input !== null && typeof input === 'object' && input.constructor === Object
}

/**
An array-like value has all the properties of an array, but is not an array instance. Examples in the `arguments` object. Returns true if the input value is an object, not null and has a `length` property with a numeric value.

@param {*} - the input to test
@returns {boolean}
@static
@example
function sum(x, y){
    console.log(t.isArrayLike(arguments))
    // prints `true`
}
*/
function isArrayLike (input) {
  return isObject(input) && typeof input.length === 'number'
}

/**
returns true if the typeof input is `"object"`, but not null!
@param {*} - the input to test
@returns {boolean}
@static
*/
function isObject (input) {
  return typeof input === 'object' && input !== null
}

/**
Returns true if the input value is defined
@param {*} - the input to test
@returns {boolean}
@static
*/
function isDefined (input) {
  return typeof input !== 'undefined'
}

/**
Returns true if the input value is a string
@param {*} - the input to test
@returns {boolean}
@static
*/
function isString (input) {
  return typeof input === 'string'
}

/**
Returns true if the input value is a boolean
@param {*} - the input to test
@returns {boolean}
@static
*/
function isBoolean (input) {
  return typeof input === 'boolean'
}

/**
Returns true if the input value is a function
@param {*} - the input to test
@returns {boolean}
@static
*/
function isFunction (input) {
  return typeof input === 'function'
}
