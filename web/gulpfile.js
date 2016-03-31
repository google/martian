//require('babel/register');

var gulp = require('gulp');
var babel = require('gulp-babel');
var mocha = require('gulp-mocha');
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
