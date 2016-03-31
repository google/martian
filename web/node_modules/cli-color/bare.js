'use strict';

var d              = require('d')
  , assign         = require('es5-ext/object/assign')
  , forEach        = require('es5-ext/object/for-each')
  , map            = require('es5-ext/object/map')
  , primitiveSet   = require('es5-ext/object/primitive-set')
  , setPrototypeOf = require('es5-ext/object/set-prototype-of')
  , includes       = require('es5-ext/string/#/contains')
  , memoize        = require('memoizee')
  , memoizeMethods = require('memoizee/methods')

  , join = Array.prototype.join, defineProperty = Object.defineProperty
  , max = Math.max, min = Math.min
  , variantModes = primitiveSet('_fg', '_bg')
  , xtermMatch, getFn;

var mods = assign({
	// Style
	bold:      { _bold: [1, 22] },
	italic:    { _italic: [3, 23] },
	underline: { _underline: [4, 24] },
	blink:     { _blink: [5, 25] },
	inverse:   { _inverse: [7, 27] },
	strike:    { _strike: [9, 29] }

	// Color
}, ['black', 'red', 'green', 'yellow', 'blue', 'magenta', 'cyan', 'white']
	.reduce(function (obj, color, index) {
		// foreground
		obj[color] = { _fg: [30 + index, 39] };
		obj[color + 'Bright'] = { _fg: [90 + index, 39] };

		// background
		obj['bg' + color[0].toUpperCase() + color.slice(1)] = { _bg: [40 + index, 49] };
		obj['bg' + color[0].toUpperCase() + color.slice(1) + 'Bright'] = { _bg: [100 + index, 49] };

		return obj;
	}, {}));

// Some use cli-color as: console.log(clc.red('Error!'));
// Which is inefficient as on each call it configures new clc object
// with memoization we reuse once created object
var memoized = memoize(function (scope, mod) {
	return defineProperty(getFn(), '_cliColorData', d(assign({}, scope._cliColorData, mod)));
});

var proto = Object.create(Function.prototype, assign(map(mods, function (mod) {
	return d.gs(function () { return memoized(this, mod); });
}), memoizeMethods({
	// xterm (255) color
	xterm: d(function (code) {
		code = isNaN(code) ? 255 : min(max(code, 0), 255);
		return defineProperty(getFn(), '_cliColorData',
			d(assign({}, this._cliColorData, {
				_fg: [xtermMatch ? xtermMatch[code] : ('38;5;' + code), 39]
			})));
	}),
	bgXterm: d(function (code) {
		code = isNaN(code) ? 255 : min(max(code, 0), 255);
		return defineProperty(getFn(), '_cliColorData',
			d(assign({}, this._cliColorData, {
				_bg: [xtermMatch ? (xtermMatch[code] + 10) : ('48;5;' + code), 49]
			})));
	})
})));

var getEndRe = memoize(function (code) {
	return new RegExp('\x1b\\[' + code + 'm', 'g');
}, { primitive: true });

if (process.platform === 'win32') xtermMatch = require('./lib/xterm-match');

getFn = function () {
	return setPrototypeOf(function self(/*…msg*/) {
		var start = '', end = '', msg = join.call(arguments, ' '), conf = self._cliColorData
		  , hasAnsi = includes.call(msg, '\x1b[');
		forEach(conf, function (mod, key) {
			end = '\x1b[' + mod[1] + 'm' + end;
			start += '\x1b[' + mod[0] + 'm';
			if (hasAnsi) {
				msg = msg.replace(getEndRe(mod[1]), variantModes[key] ? '\x1b[' + mod[0] + 'm' : '');
			}
		}, null, true);
		return start + msg + end;
	}, proto);
};

module.exports = Object.defineProperties(getFn(), {
	xtermSupported: d(!xtermMatch),
	_cliColorData: d('', {})
});
