'use strict';

var gulp = require('gulp');
const eslint = require('gulp-eslint');
var webpack = require('webpack');
var babel = require('babel-core/register');

// bundles app using webpack
gulp.task('bundle', function(done) {
    webpack(require('./webpack.config.js')).run(function(err, stats) {
        if (err) {
            console.log('Error:', err);
        } else {
            console.log(stats.toString());
        }
        done(); 
    });
});

// run ES hint 
gulp.task('lint', () => {
	return gulp.src('public/scripts')
		.pipe(eslint())
		.pipe(eslint.format())
		.pipe(eslint.failAfterError());
});

// lint and bundle by default
gulp.task('default', ['lint', 'bundle'], function() {
    gulp.watch('public/scripts/*.js', ['lint', 'bundle']);
});
