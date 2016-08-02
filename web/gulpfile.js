
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package martian provides an HTTP/1.1 proxy with an API for configurable
// request and response modifiers.

// require('babel/register');

var babel = require('gulp-babel');
var gulp = require('gulp');
var mocha = require('gulp-mocha');
var serve = require('gulp-serve');
var wct = require('web-component-tester');

wct.gulp.init(gulp);

gulp.task('js', function() {
  return gulp.src('{scripts,test}/**/*.js')
    .pipe(babel())
    .pipe(gulp.dest('dist'));
});

gulp.task('test', ['js'], function() {
  return gulp.src('dist/test/**/*-test.js')
    .pipe(mocha());
});

gulp.task('serve', serve('public'));
gulp.task('serve-build', serve(['public', 'build']));
gulp.task('serve-prod', serve({
  root: ['public', 'build'],
  port: 443,
  https: false,
  // middleware: function(req, res) {
  //   // custom optional middleware 
  // }
}));