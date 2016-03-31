/**
 * @license
 * Copyright (c) 2015 The Polymer Project Authors. All rights reserved.
 * This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
 * The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
 * The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
 * Code distributed by Google as part of the polymer project is also
 * subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
 */

// jshint node:true
'use strict';

var fs = require('fs');
var path = require('path');
var url = require('url');

var FSResolver = require('./fs-resolver');

/**
 * A single redirect configuration
 * @param {Object} config              The configuration object
 * @param {string} config.protocol     The protocol this redirect matches.
 * @param {string} config.hostname     The host name this redirect matches.
 * @param {string} config.path         The part of the path to match and
 *                                     replace with 'redirectPath'
 * @param {string} config.redirectPath The local filesystem path that should
 *                                     replace "protocol://hosname/path/"
 */
function ProtocolRedirect(config){
  this.protocol = config.protocol;
  this.hostname = config.hostname;
  this.path = config.path;
  this.redirectPath = config.redirectPath;
}

ProtocolRedirect.prototype = {
  /**
   * The protocol this redirect matches.
   * @type {string}
   */
  protocol: null,
  /**
   * The host name this redirect matches.
   * @type {string}
   */
  hostname: null,

  /**
   * The part of the path to match and replace with 'redirectPath'
   * @type {string}
   */
  path: null,

  /**
   * The local filesystem path that should replace "protocol://hosname/path/"
   * @type {string}
   */
  redirectPath: null,

  redirect: function redirect(uri) {
    var parsed = url.parse(uri);
    if (this.protocol !== parsed.protocol) {
      return null;
    } else if (this.hostname !== parsed.hostname) {
      return null;
    } else if (parsed.pathname.indexOf(this.path) !== 0) {
      return null;
    }
    return path.join(this.redirectPath,
      parsed.pathname.slice(this.path.length));
  }
};

/**
 * Resolves protocol://hostname/path to the local filesystem.
 * @constructor
 * @memberof hydrolysis
 * @param {Object} config  configuration options.
 * @param {string} config.root Filesystem root to search. Defaults to the
 *     current working directory.
 * @param {Array.<ProtocolRedirect>} redirects A list of protocol redirects
 *     for the resolver. They are checked for matching first-to-last.
 */
function RedirectResolver(config) {
  FSResolver.call(this, config);
  this.redirects = config.redirects || [];
}

RedirectResolver.prototype = Object.create(FSResolver.prototype);

RedirectResolver.prototype.accept = function(uri, deferred) {
  for (var i = 0; i < this.redirects.length; i++) {
    var redirected = this.redirects[i].redirect(uri);
    if (redirected) {
      return FSResolver.prototype.accept.call(this, redirected, deferred);
    }
  }
  return false;
};

RedirectResolver.prototype.constructor = RedirectResolver;
RedirectResolver.ProtocolRedirect = ProtocolRedirect;


module.exports = RedirectResolver;
