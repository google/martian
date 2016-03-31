/**
 * @license
 * Copyright (c) 2015 The Polymer Project Authors. All rights reserved.
 * This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
 * The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
 * The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
 * Code distributed by Google as part of the polymer project is also
 * subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
 */
// jshint node: true
'use strict';
// jshint -W079
var Promise = global.Promise || require('es6-promise').Promise;
// jshint +W079

var dom5 = require('dom5');
var url = require('url');

var docs = require('./ast-utils/docs');
var FileLoader = require('./loader/file-loader');
var importParse = require('./ast-utils/import-parse');
var jsParse = require('./ast-utils/js-parse');
var NoopResolver = require('./loader/noop-resolver');
var StringResolver = require('./loader/string-resolver');

function reduceMetadata(m1, m2) {
  return {
    elements:  m1.elements.concat(m2.elements),
    features:  m1.features.concat(m2.features),
    behaviors: m1.behaviors.concat(m2.behaviors),
  };
}

var EMPTY_METADATA = {elements: [], features: [], behaviors: []};

/**
 * Parse5's representation of a parsed html document
 * @typedef {Object} DocumentAST
 * @memberof hydrolysis
 */

/**
 * espree's representation of a parsed html document
 * @typedef {Object} JSAST
 * @memberof hydrolysis
 */

/**
 * Package of a parsed JS script
 * @typedef {Object} ParsedJS
 * @property {JSAST} ast The script's AST
 * @property {DocumentAST} scriptElement If inline, the script's containing tag.
 * @memberof hydrolysis
 */

/**
 * The metadata for a single polymer element
 * @typedef {Object} ElementDescriptor
 * @memberof hydrolysis
 */

/**
 * The metadata for a Polymer feature.
 * @typedef {Object} FeatureDescriptor
 * @memberof hydrolysis
 */

/**
 * The metadata for a Polymer behavior mixin.
 * @typedef {Object} BehaviorDescriptor
 * @memberof hydrolysis
 */

/**
 * The metadata for all features and elements defined in one document
 * @typedef {Object} DocumentDescriptor
 * @memberof hydrolysis
 * @property {Array<ElementDescriptor>} elements The elements from the document
 * @property {Array<FeatureDescriptor>}  features The features from the document
 * @property {Array<FeatureDescriptor>}  behaviors The behaviors from the document
 */

/**
 * The metadata of an entire HTML document, in promises.
 * @typedef {Object} AnalyzedDocument
 * @memberof hydrolysis
 * @property {string} href The url of the document.
 * @property {Promise<ParsedImport>}  htmlLoaded The parsed representation of
 *     the doc. Use the `ast` property to get the full `parse5` ast
 *
 * @property {Promise<Array<string>>} depsLoaded Resolves to the list of this
 *     Document's transitive import dependencies
 *
 * @property {Array<string>} depHrefs The direct dependencies of the document.
 *
 * @property {Promise<DocumentDescriptor>} metadataLoaded Resolves to the list of
 *     this Document's import dependencies
 */

/**
 * A database of Polymer metadata defined in HTML
 *
 * @constructor
 * @memberOf hydrolysis
 * @param  {boolean} attachAST  If true, attach a parse5 compliant AST
 * @param  {FileLoader=} loader An optional `FileLoader` used to load external
 *                              resources
 */
var Analyzer = function Analyzer(attachAST,
                                 loader) {
  this.loader = loader;

  /**
   * A list of all elements the `Analyzer` has metadata for.
   * @member {Array.<ElementDescriptor>}
   */
  this.elements = [];

  /**
   * A view into `elements`, keyed by tag name.
   * @member {Object.<string,ElementDescriptor>}
   */
  this.elementsByTagName = {};

  /**
   * A list of API features added to `Polymer.Base` encountered by the
   * analyzer.
   * @member {Array<FeatureDescriptor>}
   */
  this.features = [];

  /**
   * The behaviors collected by the analysis pass.
   *
   * @member {Array<BehaviorDescriptor>}
   */
  this.behaviors = [];

  /**
   * The behaviors collected by the analysis pass by name.
   *
   * @member {Object<string,BehaviorDescriptor>}
   */
  this.behaviorsByName = {};

  /**
   * A map, keyed by absolute path, of Document metadata.
   * @member {Object<string,AnalyzedDocument>}
   */
  this.html = {};

  /**
   * A map, keyed by path, of HTML document ASTs.
   * @type {Object}
   */
  this.parsedDocuments = {};

  /**
   * A map, keyed by path, of JS script ASTs.
   *
   * If the path is an HTML file with multiple scripts, the entry will be an array of scripts.
   *
   * @type {Object<string,Array<ParsedJS>>}
   */
  this.parsedScripts = {};



  /**
   * A map, keyed by path, of document content.
   * @type {Object}
   */
  this._content = {};
};

/**
 * Options for `Analyzer.analzye`
 * @typedef {Object} LoadOptions
 * @memberof hydrolysis
 * @property {boolean} noAnnotations Whether `annotate()` should be skipped.
 * @property {String=} content Content to resolve `href` to instead of loading
 *     from the file system.
 * @property {boolean} clean Whether the generated descriptors should be cleaned
 *     of redundant data.
 * @property {string=} resolver.
 *     `xhr` to use XMLHttpRequest
 *     `fs` to use the local filesystem.
 *     `permissive` to use the local filesystem and return empty files when a
 *     path can't be found.
 *     Default is `fs` in node and `xhr` in the browser.
 * @property {function(string): boolean} filter A predicate function that
 *     indicates which files should be ignored by the loader. By default all
 *     files not located under the dirname of `href` will be ignored.
 */

/**
 * Shorthand for transitively loading and processing all imports beginning at
 * `href`.
 *
 * In order to properly filter paths, `href` _must_ be an absolute URI.
 *
 * @param {string} href The root import to begin loading from.
 * @param {LoadOptions=} options Any additional options for the load.
 * @return {Promise<Analyzer>} A promise that will resolve once `href` and its
 *     dependencies have been loaded and analyzed.
 */
Analyzer.analyze = function analyze(href, options) {
  options = options || {};
  options.filter = options.filter || _defaultFilter(href);

  var loader = new FileLoader();

  var resolver = options.resolver;
  if (resolver === undefined) {
    if (typeof window === 'undefined') {
      resolver = 'fs';
    } else {
      resolver = 'xhr';
    }
  }
  var PrimaryResolver;
  if (resolver === 'fs') {
    PrimaryResolver = require('./loader/fs-resolver');
  } else if (resolver === 'xhr') {
    PrimaryResolver = require('./loader/xhr-resolver');
  } else if (resolver === 'permissive') {
    PrimaryResolver = require('./loader/error-swallowing-fs-resolver');
  } else {
    throw new Error("Resolver must be one of 'fs' or 'xhr'");
  }

  loader.addResolver(new PrimaryResolver(options));
  if (options.content) {
    loader.addResolver(new StringResolver({url: href, content: options.content}));
  }
  loader.addResolver(new NoopResolver({test: options.filter}));

  var analyzer = new this(null, loader);
  return analyzer.metadataTree(href).then(function(root) {
    if (!options.noAnnotations) {
      analyzer.annotate();
    }
    if (options.clean) {
      analyzer.clean();
    }
    return Promise.resolve(analyzer);
  });
};

/**
 * @private
 * @param {string} href
 * @return {function(string): boolean}
 */
function _defaultFilter(href) {
  // Everything up to the last `/` or `\`.
  var base = href.match(/^(.*?)[^\/\\]*$/)[1];
  return function(uri) {
    return uri.indexOf(base) !== 0;
  };
}

Analyzer.prototype.load = function load(href) {
  return this.loader.request(href).then(function(content) {
    return new Promise(function(resolve, reject) {
      setTimeout(function() {
        this._content[href] = content;
        resolve(this._parseHTML(content, href));
      }.bind(this), 0);
    }.bind(this)).catch(function(err){
      console.error("Error processing document at " + href);
      throw err;
    });
  }.bind(this));
};

/**
 * Returns an `AnalyzedDocument` representing the provided document
 * @private
 * @param  {string} htmlImport Raw text of an HTML document
 * @param  {string} href       The document's URL.
 * @return {AnalyzedDocument}       An  `AnalyzedDocument`
 */
Analyzer.prototype._parseHTML = function _parseHTML(htmlImport,
                                                  href) {
  if (href in this.html) {
    return this.html[href];
  }
  var depsLoaded = [];
  var depHrefs = [];
  var metadataLoaded = Promise.resolve(EMPTY_METADATA);
  var parsed;
  try {
    parsed = importParse(htmlImport, href);
  } catch (err) {
    console.error('Error parsing!');
    throw err;
  }
  var htmlLoaded = Promise.resolve(parsed);
  if (parsed.script) {
    metadataLoaded = this._processScripts(parsed.script, href);
  }
  var commentText = parsed.comment.map(function(comment){
    return dom5.getTextContent(comment);
  });
  var pseudoElements = docs.parsePseudoElements(commentText);
  pseudoElements.forEach(function(element){
    element.contentHref = href;
    this.elements.push(element);
    this.elementsByTagName[element.is] = element;
  }.bind(this));
  metadataLoaded = metadataLoaded.then(function(metadata){
    var metadataEntry = {
      elements: pseudoElements,
      features: [],
      behaviors: []
    };
    return [metadata, metadataEntry].reduce(reduceMetadata);
  });
  depsLoaded.push(metadataLoaded);


  if (this.loader) {
    var baseUri = href;
    if (parsed.base.length > 1) {
      console.error("Only one base tag per document!");
      throw "Multiple base tags in " + href;
    } else if (parsed.base.length == 1) {
      var baseHref = dom5.getAttribute(parsed.base[0], "href");
      if (baseHref) {
        baseHref = baseHref + "/";
        baseUri = url.resolve(baseUri, baseHref);
      }
    }
    parsed.import.forEach(function(link) {
      var linkurl = dom5.getAttribute(link, 'href');
      if (linkurl) {
        var resolvedUrl = url.resolve(baseUri, linkurl);
        depHrefs.push(resolvedUrl);
        depsLoaded.push(this._dependenciesLoadedFor(resolvedUrl, href));
      }
    }.bind(this));
    parsed.style.forEach(function(styleElement) {
      if (polymerExternalStyle(styleElement)) {
        var styleHref = dom5.getAttribute(styleElement, 'href');
        if (href) {
          styleHref = url.resolve(baseUri, styleHref);
          depsLoaded.push(this.loader.request(styleHref).then(function(content){
            this._content[styleHref] = content;
          }.bind(this)));
        }
      }
    }.bind(this));
  }
  depsLoaded = Promise.all(depsLoaded)
        .then(function() {return depHrefs;})
        .catch(function(err) {throw err;});
  this.parsedDocuments[href] = parsed.ast;
  this.html[href] = {
      href: href,
      htmlLoaded: htmlLoaded,
      metadataLoaded: metadataLoaded,
      depHrefs: depHrefs,
      depsLoaded: depsLoaded
  };
  return this.html[href];
};

Analyzer.prototype._processScripts = function _processScripts(scripts, href) {
  var scriptPromises = [];
  scripts.forEach(function(script) {
    scriptPromises.push(this._processScript(script, href));
  }.bind(this));
  return Promise.all(scriptPromises).then(function(metadataList) {
    return metadataList.reduce(reduceMetadata, EMPTY_METADATA);
  });
};

Analyzer.prototype._processScript = function _processScript(script, href) {
  var src = dom5.getAttribute(script, 'src');
  var parsedJs;
  if (!src) {
    try {
      parsedJs = jsParse((script.childNodes.length) ? script.childNodes[0].value : '');
    } catch (err) {
      // Figure out the correct line number for the error.
      var line = 0;
      var col = 0;
      if (script.__ownerDocument && script.__ownerDocument == href) {
        line = script.__locationDetail.line - 1;
        col = script.__locationDetail.column - 1;
      }
      line += err.lineNumber;
      col += err.column;
      var message = "Error parsing script in " + href + " at " + line + ":" + col;
      message += "\n" + err.stack;
      var fixedErr = new Error(message);
      fixedErr.location = {line: line, column: col};
      fixedErr.ownerDocument = script.__ownerDocument;
      return Promise.reject(fixedErr);
    }
    if (parsedJs.elements) {
      parsedJs.elements.forEach(function(element) {
        element.scriptElement = script;
        element.contentHref = href;
        this.elements.push(element);
        if (element.is in this.elementsByTagName) {
          console.warn('Ignoring duplicate element definition: ' + element.is);
        } else {
          this.elementsByTagName[element.is] = element;
        }
      }.bind(this));
    }
    if (parsedJs.features) {
      parsedJs.features.forEach(function(feature){
        feature.contentHref = href;
        feature.scriptElement = script;
      });
      this.features = this.features.concat(parsedJs.features);
    }
    if (parsedJs.behaviors) {
      parsedJs.behaviors.forEach(function(behavior){
        behavior.contentHref = href;
        this.behaviorsByName[behavior.is] = behavior;
        this.behaviorsByName[behavior.symbol] = behavior;
      }.bind(this));
      this.behaviors = this.behaviors.concat(parsedJs.behaviors);
    }
    if (!Object.hasOwnProperty.call(this.parsedScripts, href)) {
      this.parsedScripts[href] = [];
    }
    var scriptElement;
    if (script.__ownerDocument && script.__ownerDocument == href) {
      scriptElement = script;
    }
    this.parsedScripts[href].push({
      ast: parsedJs.parsedScript,
      scriptElement: scriptElement
    });
    return parsedJs;
  }
  if (this.loader) {
    var resolvedSrc = url.resolve(href, src);
    return this.loader.request(resolvedSrc).then(function(content) {
      this._content[resolvedSrc] = content;
      var resolvedScript = Object.create(script);
      resolvedScript.childNodes = [{value: content}];
      resolvedScript.attrs = resolvedScript.attrs.slice();
      dom5.removeAttribute(resolvedScript, 'src');
      return this._processScript(resolvedScript, resolvedSrc);
    }.bind(this)).catch(function(err) {throw err;});
  } else {
    return Promise.resolve(EMPTY_METADATA);
  }
};

Analyzer.prototype._dependenciesLoadedFor = function _dependenciesLoadedFor(href, root) {
  var found = {};
  if (root !== undefined) {
    found[root] = true;
  }
  return this._getDependencies(href, found).then(function(deps) {
    var depMetadataLoaded = [];
    var depPromises = deps.map(function(depHref){
      return this.load(depHref).then(function(htmlMonomer) {
        return htmlMonomer.metadataLoaded;
      });
    }.bind(this));
    return Promise.all(depPromises);
  }.bind(this));
};

/**
 * List all the html dependencies for the document at `href`.
 * @param  {string}                   href      The href to get dependencies for.
 * @param  {Object.<string,boolean>=} found     An object keyed by URL of the
 *     already resolved dependencies.
 * @param  {boolean=}                transitive Whether to load transitive
 *     dependencies. Defaults to true.
 * @return {Array.<string>}  A list of all the html dependencies.
 */
Analyzer.prototype._getDependencies = function _getDependencies(href, found, transitive) {
  if (found === undefined) {
    found = {};
    found[href] = true;
  }
  if (transitive === undefined) {
    transitive = true;
  }
  var deps = [];
  return this.load(href).then(function(htmlMonomer) {
    var transitiveDeps = [];
    htmlMonomer.depHrefs.forEach(function(depHref){
      if (found[depHref]) {
        return;
      }
      deps.push(depHref);
      found[depHref] = true;
      if (transitive) {
        transitiveDeps.push(this._getDependencies(depHref, found));
      }
    }.bind(this));
    return Promise.all(transitiveDeps);
  }.bind(this)).then(function(transitiveDeps) {
    var alldeps = transitiveDeps.reduce(function(a, b) {
      return a.concat(b);
    }, []).concat(deps);
    return alldeps;
  });
};

function matchesDocumentFolder(descriptor, href) {
  if (!descriptor.contentHref) {
    return false;
  }
  var descriptorDoc = url.parse(descriptor.contentHref);
  if (!descriptorDoc || !descriptorDoc.pathname) {
    return false;
  }
  var searchDoc = url.parse(href);
  if (!searchDoc || !searchDoc.pathname) {
    return false;
  }
  var searchPath = searchDoc.pathname;
  var lastSlash = searchPath.lastIndexOf("/");
  if (lastSlash > 0) {
    searchPath = searchPath.slice(0, lastSlash);
  }
  return descriptorDoc.pathname.indexOf(searchPath) === 0;
}

/**
 * Returns the elements defined in the folder containing `href`.
 * @param {string} href path to search.
 * @return {Array.<ElementDescriptor>}
 */
Analyzer.prototype.elementsForFolder = function elementsForFolder(href) {
  return this.elements.filter(function(element){
    return matchesDocumentFolder(element, href);
  });
};

/**
 * Returns the behaviors defined in the folder containing `href`.
 * @param {string} href path to search.
 * @return {Array.<BehaviorDescriptor>}
 */
Analyzer.prototype.behaviorsForFolder = function behaviorsForFolder(href) {
  return this.behaviors.filter(function(behavior){
    return matchesDocumentFolder(behavior, href);
  });
};

/**
 * Returns a promise that resolves to a POJO representation of the import
 * tree, in a format that maintains the ordering of the HTML imports spec.
 * @param {string} href the import to get metadata for.
 * @return {Promise}
 */
Analyzer.prototype.metadataTree = function metadataTree(href) {
  return this.load(href).then(function(monomer){
    var loadedHrefs = {};
    loadedHrefs[href] = true;
    return this._metadataTree(monomer, loadedHrefs);
  }.bind(this));
};

Analyzer.prototype._metadataTree = function _metadataTree(htmlMonomer,
                                                          loadedHrefs) {
  if (loadedHrefs === undefined) {
    loadedHrefs = {};
  }
  return htmlMonomer.metadataLoaded.then(function(metadata) {
    metadata = {
      elements: metadata.elements,
      features: metadata.features,
      href: htmlMonomer.href
    };
    return htmlMonomer.depsLoaded.then(function(hrefs) {
      var depMetadata = [];
      hrefs.forEach(function(href) {
        var metadataPromise = Promise.resolve(true);
        if (depMetadata.length > 0) {
          metadataPromise = depMetadata[depMetadata.length - 1];
        }
        metadataPromise = metadataPromise.then(function() {
          if (!loadedHrefs[href]) {
            loadedHrefs[href] = true;
            return this._metadataTree(this.html[href], loadedHrefs);
          } else {
            return Promise.resolve({});
          }
        }.bind(this));
        depMetadata.push(metadataPromise);
      }.bind(this));
      return Promise.all(depMetadata).then(function(importMetadata) {
        metadata.imports = importMetadata;
        return htmlMonomer.htmlLoaded.then(function(parsedHtml) {
          metadata.html = parsedHtml;
          if (metadata.elements) {
            metadata.elements.forEach(function(element) {
              attachDomModule(parsedHtml, element);
            });
          }
          return metadata;
        });
      });
    }.bind(this));
  }.bind(this));
};

function matchingImport(importElement) {
  var matchesTag = dom5.predicates.hasTagName(importElement.tagName);
  var matchesHref = dom5.predicates.hasAttrValue('href', dom5.getAttribute(importElement, 'href'));
  var matchesRel = dom5.predicates.hasAttrValue('rel', dom5.getAttribute(importElement, 'rel'));
  return dom5.predicates.AND(matchesTag, matchesHref, matchesRel);
}

// TODO(ajo): Refactor out of vulcanize into dom5.
var polymerExternalStyle = dom5.predicates.AND(
  dom5.predicates.hasTagName('link'),
  dom5.predicates.hasAttrValue('rel', 'import'),
  dom5.predicates.hasAttrValue('type', 'css')
);

var externalScript = dom5.predicates.AND(
  dom5.predicates.hasTagName('script'),
  dom5.predicates.hasAttr('src')
);

var isHtmlImportNode = dom5.predicates.AND(
  dom5.predicates.hasTagName('link'),
  dom5.predicates.hasAttrValue('rel', 'import'),
  dom5.predicates.NOT(
    dom5.predicates.hasAttrValue('type', 'css')
  )
);

Analyzer.prototype._inlineStyles = function _inlineStyles(ast, href) {
  var cssLinks = dom5.queryAll(ast, polymerExternalStyle);
  cssLinks.forEach(function(link) {
    var linkHref = dom5.getAttribute(link, 'href');
    var uri = url.resolve(href, linkHref);
    var content = this._content[uri];
    var style = dom5.constructors.element('style');
    dom5.setTextContent(style, '\n' + content + '\n');
    dom5.replace(link, style);
  }.bind(this));
  return cssLinks.length > 0;
};

Analyzer.prototype._inlineScripts = function _inlineScripts(ast, href) {
  var scripts = dom5.queryAll(ast, externalScript);
  scripts.forEach(function(script) {
    var scriptHref = dom5.getAttribute(script, 'src');
    var uri = url.resolve(href, scriptHref);
    var content = this._content[uri];
    var inlined = dom5.constructors.element('script');
    dom5.setTextContent(inlined, '\n' + content + '\n');
    dom5.replace(script, inlined);
  }.bind(this));
  return scripts.length > 0;
};

Analyzer.prototype._inlineImports = function _inlineImports(ast, href, loaded) {
  var imports = dom5.queryAll(ast, isHtmlImportNode);
  imports.forEach(function(htmlImport) {
    var importHref = dom5.getAttribute(htmlImport, 'href');
    var uri = url.resolve(href, importHref);
    if (loaded[uri]) {
      dom5.remove(htmlImport);
      return;
    }
    var content = this.getLoadedAst(uri, loaded);
    dom5.replace(htmlImport, content);
  }.bind(this));
  return imports.length > 0;
};

/**
 * Returns a promise resolving to a form of the AST with all links replaced
 * with the document they link to. .css and .script files become &lt;style&gt; and
 * &lt;script&gt;, respectively.
 *
 * The elements in the loaded document are unmodified from their original
 * documents.
 *
 * @param {string} href The document to load.
 * @param {Object.<string,boolean>=} loaded An object keyed by already loaded documents.
 * @return {Promise.<DocumentAST>}
 */
Analyzer.prototype.getLoadedAst = function getLoadedAst(href, loaded) {
  if (!loaded) {
    loaded = {};
  }
  loaded[href] = true;
  var parsedDocument = this.parsedDocuments[href];
  var analyzedDocument = this.html[href];
  var astCopy = dom5.parse(dom5.serialize(parsedDocument));
  // Whenever we inline something, reset inlined to true to know that anoather
  // inlining pass is needed;
  this._inlineStyles(astCopy, href);
  this._inlineScripts(astCopy, href);
  this._inlineImports(astCopy, href, loaded);
  return astCopy;
};

/**
 * Calls `dom5.nodeWalkAll` on each document that `Anayzler` has laoded.
 * @param  {Object} predicate A dom5 predicate.
 * @return {Object}
 */
Analyzer.prototype.nodeWalkDocuments = function nodeWalkDocuments(predicate) {
  for (var href in this.parsedDocuments) {
    var match = dom5.nodeWalk(this.parsedDocuments[href], predicate);
    if (match) {
      return match;
    }
  }
  return null;
};

/**
 * Calls `dom5.nodeWalkAll` on each document that `Anayzler` has laoded.
 * @param  {Object} predicate A dom5 predicate.
 * @return {Object}
 */
Analyzer.prototype.nodeWalkAllDocuments = function nodeWalkDocuments(predicate) {
  var results = [];
  for (var href in this.parsedDocuments) {
    var newNodes = dom5.nodeWalkAll(this.parsedDocuments[href], predicate);
    results = results.concat(newNodes);
  }
  return results;
};

/** Annotates all loaded metadata with its documentation. */
Analyzer.prototype.annotate = function annotate() {
  if (this.features.length > 0) {
    var featureEl = docs.featureElement(this.features);
    this.elements.unshift(featureEl);
    this.elementsByTagName[featureEl.is] = featureEl;
  }
  var behaviorsByName = this.behaviorsByName;
  var elementHelper = function(descriptor){
    docs.annotateElement(descriptor, behaviorsByName);
  };
  this.elements.forEach(elementHelper);
  this.behaviors.forEach(elementHelper); // Same shape.
  this.behaviors.forEach(function(behavior){
    if (behavior.is !== behavior.symbol && behavior.symbol) {
      this.behaviorsByName[behavior.symbol] = undefined;
    }
  }.bind(this));
};

function attachDomModule(parsedImport, element) {
  var domModules = parsedImport['dom-module'];
  for (var i = 0, domModule; i < domModules.length; i++) {
    domModule = domModules[i];
    if (dom5.getAttribute(domModule, 'id') === element.is) {
      element.domModule = domModule;
      return;
    }
  }
}

/** Removes redundant properties from the collected descriptors. */
Analyzer.prototype.clean = function clean() {
  this.elements.forEach(docs.cleanElement);
};

module.exports = Analyzer;
