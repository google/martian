# Crisper
> Split inline scripts from an HTML file for CSP compliance

## Usage

Command line usage:

```
cat index.html | crisper -h build.html -j build.js
crisper --source index.html --html build.html --js build.js
crisper --html build.html --js build.js index.html
```

The output html file will load the output js file at the top of `<head>` with a `<script defer>` element.

Optional Flags:

  - `--script-in-head=false`
    - in the output HTML file, place the script at the end of `<body>`
    - **Note**: Only use this if you need `document.write` support.
  - `--only-split`
    - Do not write include a `<script>` tag in the output HTML
      file
  - `--always-write-script`
    - Always create a .js file, even without any `<script>`
      elements.

Library usage:

```js
var output = crisper({
  source: 'source HTML string',
  jsFileName: 'output js file name.js',
  scriptInHead: true, //default true
  onlySplit: false, // default false
  alwaysWriteScript: false // default false
});
fs.writeFile(htmlOutputFileName, output.html, 'utf-8', ...);
fs.writeFile(jsOutputFileName, output.js, 'utf-8', ...);
```

## Usage with Vulcanize

When using [vulcanize](https://github.com/Polymer/vulcanize), crisper can handle
the html string output directly and write the CSP seperated files on the command
line

```
vulcanize index.html --inline-script | crisper --html build.html --js build.js
```

Or programmatically

```js
vulcanize.process('index.html', function(err, cb) {
  if (err) {
    return cb(err);
  } else {
    var out = crisper({
      source: html,
      jsFileName: 'name of js file.js',
      scriptInHead: true, // default true
      onlySplit: false, // default false
      alwaysWriteScript: false //default false
    })
    cb(null, out.html, out.js);
  }
});
```

## Breaking Changes from 1.x
- Deprecated `split` API was removed
  - `require('crisper').split()`
- Default value of `script-in-head` flag changed to true
  - This improves load performance by parallelizing HTML and script parsing
  - This will break `document.write` calls
  - If you experience problems, you can use `--script-in-head=false` argument or
      `scriptInHead: false` in library usage.

## Build Tools

- [gulp-crisper](https://npmjs.com/package/gulp-crisper)
- [grunt-crisper](https://www.npmjs.com/package/grunt-crisper)
