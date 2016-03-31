[![view on npm](http://img.shields.io/npm/v/feature-detect-es6.svg)](https://www.npmjs.org/package/feature-detect-es6)
[![npm module downloads per month](http://img.shields.io/npm/dm/feature-detect-es6.svg)](https://www.npmjs.org/package/feature-detect-es6)
[![Build Status](https://travis-ci.org/75lb/feature-detect-es6.svg?branch=master)](https://travis-ci.org/75lb/feature-detect-es6)
[![Dependency Status](https://david-dm.org/75lb/feature-detect-es6.svg)](https://david-dm.org/75lb/feature-detect-es6)
[![js-standard-style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://github.com/feross/standard)

<a name="module_feature-detect-es6"></a>
## feature-detect-es6
Detect which ES6 features are available.

**Example**  
```js
var detect = require('feature-detect-es6')

if (detect.class() && detect.arrowFunction()){
  // safe to run ES6 code natively..
} else {
  // run your transpiled ES5..
}
```

* [feature-detect-es6](#module_feature-detect-es6)
  * [.class()](#module_feature-detect-es6.class) ⇒ <code>boolean</code>
  * [.arrowFunction()](#module_feature-detect-es6.arrowFunction) ⇒ <code>boolean</code>
  * [.let()](#module_feature-detect-es6.let) ⇒ <code>boolean</code>
  * [.const()](#module_feature-detect-es6.const) ⇒ <code>boolean</code>
  * [.newArrayFeatures()](#module_feature-detect-es6.newArrayFeatures) ⇒ <code>boolean</code>
  * [.collections()](#module_feature-detect-es6.collections) ⇒ <code>boolean</code>
  * [.generators()](#module_feature-detect-es6.generators) ⇒ <code>boolean</code>
  * [.promises()](#module_feature-detect-es6.promises) ⇒ <code>boolean</code>
  * [.templateStrings()](#module_feature-detect-es6.templateStrings) ⇒ <code>boolean</code>
  * [.symbols()](#module_feature-detect-es6.symbols) ⇒ <code>boolean</code>
  * [.destructuring()](#module_feature-detect-es6.destructuring) ⇒ <code>boolean</code>
  * [.spread()](#module_feature-detect-es6.spread) ⇒ <code>boolean</code>

<a name="module_feature-detect-es6.class"></a>
### detect.class() ⇒ <code>boolean</code>
Returns true if the `class` statement is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.arrowFunction"></a>
### detect.arrowFunction() ⇒ <code>boolean</code>
Returns true if the arrow functions available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.let"></a>
### detect.let() ⇒ <code>boolean</code>
Returns true if the `let` statement is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.const"></a>
### detect.const() ⇒ <code>boolean</code>
Returns true if the `const` statement is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.newArrayFeatures"></a>
### detect.newArrayFeatures() ⇒ <code>boolean</code>
Returns true if the [new Array features](http://exploringjs.com/es6/ch_arrays.html) are available (exluding `Array.prototype.values` which has zero support anywhere).

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.collections"></a>
### detect.collections() ⇒ <code>boolean</code>
Returns true if `Map`, `WeakMap`, `Set` and `WeakSet` are available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.generators"></a>
### detect.generators() ⇒ <code>boolean</code>
Returns true if generators are available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.promises"></a>
### detect.promises() ⇒ <code>boolean</code>
Returns true if `Promise` is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.templateStrings"></a>
### detect.templateStrings() ⇒ <code>boolean</code>
Returns true if template strings are available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.symbols"></a>
### detect.symbols() ⇒ <code>boolean</code>
Returns true if `Symbol` is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.destructuring"></a>
### detect.destructuring() ⇒ <code>boolean</code>
Returns true if destructuring is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  
<a name="module_feature-detect-es6.spread"></a>
### detect.spread() ⇒ <code>boolean</code>
Returns true if the spread operator (`...`) is available.

**Kind**: static method of <code>[feature-detect-es6](#module_feature-detect-es6)</code>  

* * *

&copy; 2015 Lloyd Brookes \<75pound@gmail.com\>. Documented by [jsdoc-to-markdown](https://github.com/jsdoc2md/jsdoc-to-markdown).
