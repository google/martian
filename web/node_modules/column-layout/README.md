[![view on npm](http://img.shields.io/npm/v/column-layout.svg)](https://www.npmjs.org/package/column-layout)
[![npm module downloads](http://img.shields.io/npm/dt/column-layout.svg)](https://www.npmjs.org/package/column-layout)
[![Build Status](https://travis-ci.org/75lb/column-layout.svg?branch=master)](https://travis-ci.org/75lb/column-layout)
[![Dependency Status](https://david-dm.org/75lb/column-layout.svg)](https://david-dm.org/75lb/column-layout)
[![js-standard-style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://github.com/feross/standard)

# column-layout
Pretty-print text data in columns.

## Synopsis
Say you have some data:
```json
[
    {
      "column 1": "The Kingdom of Scotland was a state in north-west Europe traditionally said to have been founded in 843, which joined with the Kingdom of England to form a unified Kingdom of Great Britain in 1707. Its territories expanded and shrank, but it came to occupy the northern third of the island of Great Britain, sharing a land border to the south with the Kingdom of England. ",
      "column 2": "Operation Barbarossa (German: Unternehmen Barbarossa) was the code name for Nazi Germany's invasion of the Soviet Union during World War II, which began on 22 June 1941. Over the course of the operation, about four million soldiers of the Axis powers invaded Soviet Russia along a 2,900 kilometer front, the largest invasion force in the history of warfare. In addition to troops, the Germans employed some 600,000 motor vehicles and between 600–700,000 horses."
    }
]
```

pipe it through `column-layout`:
```sh
$ cat example/two-columns.json | column-layout
```

to get this:
```
The Kingdom of Scotland was a state in     Operation Barbarossa (German: Unternehmen
north-west Europe traditionally said to    Barbarossa) was the code name for Nazi
have been founded in 843, which joined     Germany's invasion of the Soviet Union
with the Kingdom of England to form a      during World War II, which began on 22
unified Kingdom of Great Britain in 1707.  June 1941. Over the course of the
Its territories expanded and shrank, but   operation, about four million soldiers of
it came to occupy the northern third of    the Axis powers invaded Soviet Russia
the island of Great Britain, sharing a     along a 2,900 kilometer front, the
land border to the south with the Kingdom  largest invasion force in the history of
of England.                                warfare. In addition to troops, the
                                           Germans employed some 600,000 motor
                                           vehicles and between 600–700,000 horses.
```

Columns containing wrappable data are auto-sized by default to fit the available space. You can set specific widths using `--width`

```sh
$ cat example/two-columns.json | column-layout --width "column 2: 55"
```

```
The Kingdom of Scotland was a  Operation Barbarossa (German: Unternehmen Barbarossa)
state in north-west Europe     was the code name for Nazi Germany's invasion of the
traditionally said to have     Soviet Union during World War II, which began on 22
been founded in 843, which     June 1941. Over the course of the operation, about
joined with the Kingdom of     four million soldiers of the Axis powers invaded
England to form a unified      Soviet Russia along a 2,900 kilometer front, the
Kingdom of Great Britain in    largest invasion force in the history of warfare. In
1707. Its territories          addition to troops, the Germans employed some 600,000
expanded and shrank, but it    motor vehicles and between 600–700,000 horses.
came to occupy the northern
third of the island of Great
Britain, sharing a land
border to the south with the
Kingdom of England.
```

## More Examples
Read the latest npm issues: (example requires [jq](https://stedolan.github.io/jq/))
```sh
$ curl -s https://api.github.com/repos/npm/npm/issues \
| jq 'map({ number, title, login:.user.login, comments })' \
| column-layout
```
```
10263  npm run start                                            Slepperpon        4
10262  npm-shrinkwrap.json being ignored for a dependency of a  maxkorp           0
      dependency (2.14.9, 3.3.10)
10261  EPROTO Error Installing Packages                         azkaiart          2
10260  ENOENT during npm install with npm v3.3.6/v3.3.12 and    lencioni          2
      node v5.0.0
10259  npm install failed                                       geraldvillorente  1
10258  npm moves common dependencies under a dependency on      trygveaa          2
      install
10257  [NPM3] Missing top level dependencies after npm install  naholyr           0
10256  Yo meanjs app creation problem                           nrjkumar41        0
10254  sapnwrfc is not installing                               RamprasathS       0
10253  npm install deep dependence folder "node_modules"        duyetvv           2
10251  cannot npm login                                         w0ps              2
10250  Update npm-team.md                                       louislarry        0
10248  cant install module I created                            nousacademy       4
10247  Cannot install passlib                                   nicola883         3
10246  Error installing Gulp                                    AlanIsrael0       1
10245  cannot install packages through NPM                      RoyGeagea         11
10244  Remove arguments from npm-dedupe.md                      bengotow          0
 etc.
 etc.
```

## Install
As a library:

```
$ npm install column-layout --save
```

As a command-line tool:
```
$ npm install -g column-layout
```

## API Reference

* [column-layout](#module_column-layout)
  * [columnLayout(data, [options])](#exp_module_column-layout--columnLayout) ⇒ <code>string</code> ⏏
    * _static_
      * [.lines()](#module_column-layout--columnLayout.lines) ⇒ <code>Array</code>
    * _inner_
      * [~columnOption](#module_column-layout--columnLayout..columnOption)

<a name="exp_module_column-layout--columnLayout"></a>
### columnLayout(data, [options]) ⇒ <code>string</code> ⏏
Returns JSON data formatted in columns.

**Kind**: Exported function  

| Param | Type | Description |
| --- | --- | --- |
| data | <code>array</code> | input data |
| [options] | <code>object</code> | optional settings |
| [options.viewWidth] | <code>number</code> | maximum width of layout |
| [options.nowrap] | <code>boolean</code> | disable wrapping on all columns |
| [options.break] | <code>boolean</code> | enable word-breaking on all columns |
| [options.columns] | <code>[columnOption](#module_column-layout--columnLayout..columnOption)</code> | array of column options |
| [options.padding] | <code>object</code> | Padding values to set on each column. Per-column overrides can be set in the `options.columns` array. |
| [options.padding.left] | <code>string</code> |  |
| [options.padding.right] | <code>string</code> |  |

**Example**  
```js
> columnFormat = require("column-format")
> jsonData = [{
    col1: "Some text you wish to read in column layout",
    col2: "And some more text in column two. "
}]
> columnFormat(jsonData, { viewWidth: 30 })
' Some text you  And some more \n wish to read   text in       \n in column      column two.   \n layout                       \n'
```
<a name="module_column-layout--columnLayout.lines"></a>
#### columnLayout.lines() ⇒ <code>Array</code>
Identical to [column-layout](#module_column-layout) with the exception of the rendered result being returned as an array of lines, rather that a single string.

**Kind**: static method of <code>[columnLayout](#exp_module_column-layout--columnLayout)</code>  
**Example**  
```js
> columnFormat = require("column-format")
> jsonData = [{
     col1: "Some text you wish to read in column layout",
     col2: "And some more text in column two. "
}]
> columnFormat.lines(jsonData, { viewWidth: 30 })
[ ' Some text you  And some more ',
' wish to read   text in       ',
' in column      column two.   ',
' layout                       ' ]
```
<a name="module_column-layout--columnLayout..columnOption"></a>
#### columnLayout~columnOption
**Kind**: inner typedef of <code>[columnLayout](#exp_module_column-layout--columnLayout)</code>  
**Properties**

| Name | Type | Description |
| --- | --- | --- |
| width | <code>number</code> | column width |
| minWidth | <code>number</code> | column min width |
| maxWidth | <code>number</code> | column max width |
| nowrap | <code>boolean</code> | disable wrapping for this column |
| break | <code>boolean</code> | enable word-breaking for this columns |
| padding | <code>object</code> | padding options |
| padding.left | <code>string</code> | a string to pad the left of each cell (default: `" "`) |
| padding.right | <code>string</code> | a string to pad the right of each cell (default: `" "`) |


* * *

&copy; 2015 Lloyd Brookes \<75pound@gmail.com\>. Documented by [jsdoc-to-markdown](https://github.com/jsdoc2md/jsdoc-to-markdown).
