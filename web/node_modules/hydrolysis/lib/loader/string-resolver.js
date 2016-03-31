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

/**
 * A resolver that resolves to `config.content` any uri matching config.
 * @constructor
 * @memberof hydrolysis
 * @param {string|RegExp} config.url     The url or rejex to accept.
 * @param {string} config.content The content to serve for `url`.
 */
function StringResolver(config) {
  this.url = config.url;
  this.content = config.content;
  if (!this.url || !this.content) {
    throw new Error("Must provide a url and content to the string resolver.");
  }
}

StringResolver.prototype = {

  /**
   * @param {string}    uri      The absolute URI being requested.
   * @param {!Deferred} deferred The deferred promise that should be resolved if
   *     this resolver handles the URI.
   * @return {boolean} Whether the URI is handled by this resolver.
   */
  accept: function(uri, deferred) {
    if (this.url.test) {
      // this.url is a regex
      if (!this.url.test(uri)) {
        return false;
      }
    } else {
      // this.url is a string
      if (uri.search(this.url) == -1) {
        return false;
      }
    }
    deferred.resolve(this.content);
    return true;
  }
};

module.exports = StringResolver;
