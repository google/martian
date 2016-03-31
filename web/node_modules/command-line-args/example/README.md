# Examples
Most of these example files are modules exporting an array of [Option Definitions](https://github.com/75lb/command-line-args#optiondefinition-). They are consumed using the command-line-args test harness.

## Install
Install the test harness:

```
$ npm install -g command-line-args
```

## Usage
Try one of the examples out

```
$ cat example/typical.js | command-line-args --timeout 100 --src lib/*

{ timeout: 100,
  src:
   [ 'lib/argv.js',
     'lib/command-line-args.js',
     'lib/definition.js',
     'lib/definitions.js',
     'lib/option.js' ] }
```

# Validation
command-line-args parses the command line but does not validate the values received. There's one example ([validate.js](https://github.com/75lb/command-line-args/blob/master/example/validate.js)) suggesting how this could be done:

```
$ node example/validate.js
```
