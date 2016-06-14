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