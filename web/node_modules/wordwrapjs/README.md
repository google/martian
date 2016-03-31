[![view on npm](http://img.shields.io/npm/v/wordwrapjs.svg)](https://www.npmjs.org/package/wordwrapjs)
[![npm module downloads](http://img.shields.io/npm/dt/wordwrapjs.svg)](https://www.npmjs.org/package/wordwrapjs)
[![Build Status](https://travis-ci.org/75lb/wordwrapjs.svg?branch=master)](https://travis-ci.org/75lb/wordwrapjs)
[![Dependency Status](https://david-dm.org/75lb/wordwrapjs.svg)](https://david-dm.org/75lb/wordwrapjs)
[![js-standard-style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://github.com/feross/standard)

<a name="module_wordwrapjs"></a>
## wordwrapjs
Word wrapping, with a few features.

- multilingual - wraps any language using whitespace word separation.
- force-break option
- ignore pattern option (e.g. ansi escape sequences)
- wraps hypenated words

**Example**  
Wrap some sick bars in a 20 character column.

```js
> wrap = require("wordwrapjs")

> bars = "I'm rapping. I'm rapping. I'm rap rap rapping. I'm rap rap rap rap rappity rapping."
> result = wrap(bars, { width: 20 })
```

`result` now looks like this:
```
I'm rapping. I'm
rapping. I'm rap rap
rapping. I'm rap rap
rap rap rappity
rapping.
```

By default, long words will not break. Unless you insist.
```js
> url = "https://github.com/75lb/wordwrapjs"

> wrap.lines(url, { width: 18 })
[ 'https://github.com/75lb/wordwrapjs' ]

> wrap.lines(url, { width: 18, break: true })
[ 'https://github.com', '/75lb/wordwrapjs' ]
```

* [wordwrapjs](#module_wordwrapjs)
  * [wrap(text, [options])](#exp_module_wordwrapjs--wrap) ⇒ <code>string</code> ⏏
    * [.lines(text, [options])](#module_wordwrapjs--wrap.lines) ⇒ <code>Array</code>
    * [.isWrappable(text)](#module_wordwrapjs--wrap.isWrappable) ⇒ <code>boolean</code>
    * [.getWords(text)](#module_wordwrapjs--wrap.getWords) ⇒ <code>Array.&lt;string&gt;</code>

<a name="exp_module_wordwrapjs--wrap"></a>
### wrap(text, [options]) ⇒ <code>string</code> ⏏
**Kind**: Exported function  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| text | <code>string</code> |  | the input text to wrap |
| [options] | <code>object</code> |  | optional config |
| [options.width] | <code>number</code> | <code>30</code> | the max column width in characters |
| [options.ignore] | <code>RegExp</code> &#124; <code>Array.&lt;RegExp&gt;</code> |  | one or more patterns to be ignored when sizing the newly wrapped lines. For example `ignore: /\u001b.*?m/g` will ignore unprintable ansi escape sequences. |
| [options.break] | <code>boolean</code> |  | if true, words exceeding the specified `width` will be forcefully broken |
| [options.eol] | <code>string</code> | <code>&quot;os.EOL&quot;</code> | the desired new line character to use, defaults to [os.EOL](https://nodejs.org/api/os.html#os_os_eol). |

<a name="module_wordwrapjs--wrap.lines"></a>
#### wrap.lines(text, [options]) ⇒ <code>Array</code>
returns the wrapped output as an array of lines, rather than a single string

**Kind**: static method of <code>[wrap](#exp_module_wordwrapjs--wrap)</code>  

| Param | Type | Description |
| --- | --- | --- |
| text | <code>string</code> | the input text to wrap |
| [options] | <code>object</code> | same options as [wrap](#module_wordwrapjs) |

**Example**  
```js
> bars = "I'm rapping. I'm rapping. I'm rap rap rapping. I'm rap rap rap rap rappity rapping."
> wrap.lines(bars)
[ "I'm rapping. I'm rapping. I'm",
  "rap rap rapping. I'm rap rap",
  "rap rap rappity rapping." ]
```
<a name="module_wordwrapjs--wrap.isWrappable"></a>
#### wrap.isWrappable(text) ⇒ <code>boolean</code>
Returns true if the input text is wrappable

**Kind**: static method of <code>[wrap](#exp_module_wordwrapjs--wrap)</code>  

| Param | Type | Description |
| --- | --- | --- |
| text | <code>string</code> | input text |

<a name="module_wordwrapjs--wrap.getWords"></a>
#### wrap.getWords(text) ⇒ <code>Array.&lt;string&gt;</code>
Splits the input text returning an array of words

**Kind**: static method of <code>[wrap](#exp_module_wordwrapjs--wrap)</code>  

| Param | Type | Description |
| --- | --- | --- |
| text | <code>string</code> | input text |


* * *

&copy; 2015 Lloyd Brookes \<75pound@gmail.com\>. Documented by [jsdoc-to-markdown](https://github.com/jsdoc2md/jsdoc-to-markdown).
