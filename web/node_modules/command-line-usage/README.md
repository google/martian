[![view on npm](http://img.shields.io/npm/v/command-line-usage.svg)](https://www.npmjs.org/package/command-line-usage)
[![npm module downloads](http://img.shields.io/npm/dt/command-line-usage.svg)](https://www.npmjs.org/package/command-line-usage)
[![Build Status](https://travis-ci.org/75lb/command-line-usage.svg?branch=master)](https://travis-ci.org/75lb/command-line-usage)
[![Dependency Status](https://david-dm.org/75lb/command-line-usage.svg)](https://david-dm.org/75lb/command-line-usage)
[![js-standard-style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://github.com/feross/standard)

# command-line-usage
A simple template to create a usage guide. It was extracted from  [command-line-args](https://github.com/75lb/command-line-args) to faciliate arbitrary use.

```js
var getUsage = require("command-line-usage");
var usage = getUsage(definitions, options)
```

Inline ansi formatting can be used anywhere within the usage template using the formatting syntax described [here](https://github.com/75lb/ansi-escape-sequences#module_ansi-escape-sequences.format).

## Examples

### Simple
A `description` field is added to each option definition. A `title`, `description` and simple `footer` are set in the getUsage options. [Code](https://github.com/75lb/command-line-usage/blob/master/example/simple.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/simple.png)

### Groups
Demonstrates breaking the options up into groups. This example also sets a `typeLabel` on each option definition (e.g. a `typeLabel` value of `files` is more meaningful than the default `string[]`). [Code](https://github.com/75lb/command-line-usage/blob/master/example/groups.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/groups.png)

### Header
Here, the `title` is replaced with a `header` banner. This example also adds a `synopsis` list. [Code](https://github.com/75lb/command-line-usage/blob/master/example/header.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/header.png)

### Footer
The footer is displayed at the end of the template. [Code](https://github.com/75lb/command-line-usage/blob/master/example/footer.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/footer.png)

### Examples (column layout)
A list of `examples` is added. In this case the example list is defined as an array of objects (each with consistently named properties) so will be formatted by [column-layout](https://github.com/75lb/column-layout).   [Code](https://github.com/75lb/command-line-usage/blob/master/example/examples.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/example-columns.png)

### Description (column layout)
Demonstrates usage of custom column layout in the description. In this case the second column (containing the hammer and sickle) has `nowrap` enabled, as the input is already formatted as desired. [Code](https://github.com/75lb/command-line-usage/blob/master/example/description-columns.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/description-columns.png)

### Custom
Demonstrates a custom template. The `getUsage.optionList()` method exists for users that want the option list and nothing else. [Code](https://github.com/75lb/command-line-usage/blob/master/example/custom.js).

![usage](https://raw.githubusercontent.com/75lb/command-line-usage/master/example/screens/custom.png)

# API Reference

* [command-line-usage](#module_command-line-usage)
  * [getUsage(definitions, options)](#exp_module_command-line-usage--getUsage) ⇒ <code>string</code> ⏏
    * [.optionList(definitions, [group])](#module_command-line-usage--getUsage.optionList) ⇒ <code>Array.&lt;string&gt;</code>

<a name="exp_module_command-line-usage--getUsage"></a>
### getUsage(definitions, options) ⇒ <code>string</code> ⏏
**Kind**: Exported function  
<table>
  <thead>
    <tr>
      <th>Param</th><th>Type</th><th>Description</th>
    </tr>
  </thead>
  <tbody>
<tr>
    <td>definitions</td><td><code>Array.&lt;optionDefinition&gt;</code></td><td><p>an array of <a href="https://github.com/75lb/command-line-args#exp_module_definition--OptionDefinition">option definition</a> objects. In addition to the regular definition properties, command-line-usage will look for:</p>
<ul>
<li><code>description</code> - a string describing the option.</li>
<li><code>typeLabel</code> - a string to replace the default type string (e.g. <code>&lt;string&gt;</code>). It&#39;s often more useful to set a more descriptive type label, like <code>&lt;ms&gt;</code>, <code>&lt;files&gt;</code>, <code>&lt;command&gt;</code> etc.</li>
</ul>
</td>
    </tr><tr>
    <td>options</td><td><code><a href="#module_usage-options">usage-options</a></code></td><td><p>see <a href="#exp_module_usage-options--UsageOptions">UsageOptions</a>.</p>
</td>
    </tr>  </tbody>
</table>

<a name="module_command-line-usage--getUsage.optionList"></a>
#### getUsage.optionList(definitions, [group]) ⇒ <code>Array.&lt;string&gt;</code>
A helper for getting a column-format list of options and descriptions. Useful for inserting into a custom usage template.

**Kind**: static method of <code>[getUsage](#exp_module_command-line-usage--getUsage)</code>  
<table>
  <thead>
    <tr>
      <th>Param</th><th>Type</th><th>Description</th>
    </tr>
  </thead>
  <tbody>
<tr>
    <td>definitions</td><td><code>Array.&lt;optionDefinition&gt;</code></td><td><p>the definitions to Display</p>
</td>
    </tr><tr>
    <td>[group]</td><td><code>string</code></td><td><p>if specified, will output the options in this group. The special group <code>&#39;_none&#39;</code> will return options without a group specified.</p>
</td>
    </tr>  </tbody>
</table>


<a name="exp_module_usage-options--UsageOptions"></a>
## UsageOptions ⏏
The class describes all valid options for the `getUsage` function. Inline formatting can be used within any text string supplied using valid [ansi-escape-sequences formatting syntax](https://github.com/75lb/ansi-escape-sequences#module_ansi-escape-sequences.format).

**Kind**: Exported class  
* [UsageOptions](#exp_module_usage-options--UsageOptions) ⏏
  * _instance_
    * [.header](#module_usage-options--UsageOptions+header) : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
    * [.title](#module_usage-options--UsageOptions+title) : <code>string</code>
    * [.description](#module_usage-options--UsageOptions+description) : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
    * [.synopsis](#module_usage-options--UsageOptions+synopsis) : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
    * [.groups](#module_usage-options--UsageOptions+groups) : <code>object</code>
    * [.examples](#module_usage-options--UsageOptions+examples) : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
    * [.footer](#module_usage-options--UsageOptions+footer) : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
    * [.hide](#module_usage-options--UsageOptions+hide) : <code>string</code> &#124; <code>Array.&lt;string&gt;</code>
  * _inner_
    * [~textBlock](#module_usage-options--UsageOptions..textBlock) : <code>string</code> &#124; <code>Array.&lt;string&gt;</code> &#124; <code>Array.&lt;object&gt;</code> &#124; <code>Object</code>

<a name="module_usage-options--UsageOptions+header"></a>
### options.header : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
Use this field to display a banner or header above the main body.

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
<a name="module_usage-options--UsageOptions+title"></a>
### options.title : <code>string</code>
The title line at the top of the usage, typically the name of the app. By default it is underlined but this formatting can be overridden by passing a [module:usage-options~textObject](module:usage-options~textObject).

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
**Example**  
```js
{ title: "my-app" }
```
<a name="module_usage-options--UsageOptions+description"></a>
### options.description : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
A description to go underneath the title. For example, some words about what the app is for.

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
<a name="module_usage-options--UsageOptions+synopsis"></a>
### options.synopsis : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
An array of strings highlighting the main usage forms of the app.

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
<a name="module_usage-options--UsageOptions+groups"></a>
### options.groups : <code>object</code>
Specify which groups to display in the output by supplying an object of key/value pairs, where the key is the name of the group to include and the value is a string or textObject. If the value is a string it is used as the group title. Alternatively supply an object containing a `title` and `description` string.

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
**Example**  
```js
{
    main: {
        title: "Main options",
        description: "This group contains the most important options."
    },
    misc: "Miscellaneous"
}
```
<a name="module_usage-options--UsageOptions+examples"></a>
### options.examples : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
Examples

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
<a name="module_usage-options--UsageOptions+footer"></a>
### options.footer : <code>[textBlock](#module_usage-options--UsageOptions..textBlock)</code>
Displayed at the foot of the usage output.

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
**Example**  
```js
{
    footer: "Project home: [underline]{https://github.com/me/my-app}"
}
```
<a name="module_usage-options--UsageOptions+hide"></a>
### options.hide : <code>string</code> &#124; <code>Array.&lt;string&gt;</code>
If you want to hide certain options from the output, specify their names here. This is sometimes used to hide the `defaultOption`.

**Kind**: instance property of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  
**Example**  
```js
{
    hide: "files"
}
```
<a name="module_usage-options--UsageOptions..textBlock"></a>
### options~textBlock : <code>string</code> &#124; <code>Array.&lt;string&gt;</code> &#124; <code>Array.&lt;object&gt;</code> &#124; <code>Object</code>
A text block can be a string:

```js
{
  description: 'This is a single-line description.'
}
```
.. or multiple strings:
```js
{
  description: [
    'This is a multi-line description.',
    'A new string in the array represents a new line.'
  ]
}
```
.. or an array of objects. In which case, it will be formatted by [column-layout](https://github.com/75lb/column-layout):
```js
{
  description: {
    column1: 'This will go in column 1.',
    column2: 'Second column text.'
  }
}
```
If you want set specific column-layout options, pass an object with two properties: `options` and `data`.
```js
{
  description: {
    options: {
      columns: [
        { name: 'two', width: 40, nowrap: true }
      ]
    },
    data: {
      column1: 'This will go in column 1.',
      column2: 'Second column text.'
    }
  }
}
```

**Kind**: inner typedef of <code>[UsageOptions](#exp_module_usage-options--UsageOptions)</code>  


* * *

&copy; 2015 Lloyd Brookes \<75pound@gmail.com\>. Documented by [jsdoc-to-markdown](https://github.com/75lb/jsdoc-to-markdown).
