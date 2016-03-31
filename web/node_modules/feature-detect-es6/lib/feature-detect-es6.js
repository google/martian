'use strict'

/**
 * Detect which ES6 features are available.
 *
 * @module feature-detect-es6
 * @typicalname detect
 * @example
 * var detect = require('feature-detect-es6')
 *
 * if (detect.class() && detect.arrowFunction()){
 *   // safe to run ES6 code natively..
 * } else {
 *   // run your transpiled ES5..
 * }
 */

/**
 * Returns true if the `class` statement is available.
 *
 * @returns {boolean}
 */
exports.class = function () {
  return evaluates('class Something {}')
}

/**
 * Returns true if the arrow functions available.
 *
 * @returns {boolean}
 */
exports.arrowFunction = function () {
  return evaluates('var f = x => 1')
}

/**
 * Returns true if the `let` statement is available.
 *
 * @returns {boolean}
 */
exports.let = function () {
  return evaluates('let a = 1')
}

/**
 * Returns true if the `const` statement is available.
 *
 * @returns {boolean}
 */
exports.const = function () {
  return evaluates('const a = 1')
}

/**
 * Returns true if the [new Array features](http://exploringjs.com/es6/ch_arrays.html) are available (exluding `Array.prototype.values` which has zero support anywhere).
 *
 * @returns {boolean}
 */
exports.newArrayFeatures = function () {
  return typeof Array.prototype.find !== 'undefined' &&
    typeof Array.prototype.findIndex !== 'undefined' &&
    typeof Array.from !== 'undefined' &&
    typeof Array.of !== 'undefined' &&
    typeof Array.prototype.entries !== 'undefined' &&
    typeof Array.prototype.keys !== 'undefined' &&
    typeof Array.prototype.copyWithin !== 'undefined' &&
    typeof Array.prototype.fill !== 'undefined'
}

/**
 * Returns true if `Map`, `WeakMap`, `Set` and `WeakSet` are available.
 *
 * @returns {boolean}
 */
exports.collections = function () {
  return typeof Map !== 'undefined' &&
    typeof WeakMap !== 'undefined' &&
    typeof Set !== 'undefined' &&
    typeof WeakSet !== 'undefined'
}

/**
 * Returns true if generators are available.
 *
 * @returns {boolean}
 */
exports.generators = function () {
  return evaluates('function* test() {}')
}

/**
 * Returns true if `Promise` is available.
 *
 * @returns {boolean}
 */
exports.promises = function () {
  return typeof Promise !== 'undefined'
}

/**
 * Returns true if template strings are available.
 *
 * @returns {boolean}
 */
exports.templateStrings = function () {
  return evaluates('var a = `a`')
}

/**
 * Returns true if `Symbol` is available.
 *
 * @returns {boolean}
 */
exports.symbols = function () {
  return typeof Symbol !== 'undefined'
}

/**
 * Returns true if destructuring is available.
 *
 * @returns {boolean}
 */
exports.destructuring = function () {
  return evaluates("var { first: f, last: l } = { first: 'Jane', last: 'Doe' }")
}

/**
 * Returns true if the spread operator (`...`) is available.
 *
 * @returns {boolean}
 */
exports.spread = function () {
  return evaluates('Math.max(...[ 5, 10 ])')
}

function evaluates (statement) {
  try {
    eval(statement)
    return true
  } catch (err) {
    return false
  }
}
