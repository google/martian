# Service Worker Toolbox

[![Build Status](https://travis-ci.org/GoogleChrome/sw-toolbox.svg?branch=master)](https://travis-ci.org/GoogleChrome/sw-toolbox) [![Dependency Status](https://david-dm.org/googlechrome/sw-toolbox.svg)](https://david-dm.org/googlechrome/sw-toolbox) [![devDependency Status](https://david-dm.org/googlechrome/sw-toolbox/dev-status.svg)](https://david-dm.org/googlechrome/sw-toolbox#info=devDependencies)

> A collection of tools for [service workers](https://slightlyoff.github.io/ServiceWorker/spec/service_worker/)

Service Worker Toolbox provides some simple helpers for use in creating your own service workers. Specifically, it provides common caching patterns and an [expressive approach](https://googlechrome.github.io/sw-toolbox/docs/master/tutorial-api#expressive-approach) to using those strategies for runtime requests. If you're not sure what service workers are or what they are for, start with [the explainer doc](https://github.com/slightlyoff/ServiceWorker/blob/master/explainer.md).

## Install

Service Worker Toolbox is available through Bower, npm or direct from GitHub:

`bower install --save sw-toolbox`

`npm install --save sw-toolbox`

`git clone https://github.com/GoogleChrome/sw-toolbox.git`

### Register your service worker

From your registering page, register your service worker in the normal way. For example:

```javascript
navigator.serviceWorker.register('my-service-worker.js');
```
As implemented in Chrome 40 or later, a service worker must exist at the root of the scope that you intend it to control, or higher. So if you want all of the pages under `/myapp/` to be controlled by the worker, the worker script itself must be served from either `/` or `/myapp/`. The default scope is the containing path of the service worker script.

For even lower friction you can instead include the Service Worker Toolbox companion script in your HTML as shown below. Be aware that this is not customizable. If you need to do anything fancier than registering with a default scope, you'll need to use the standard registration.

```html
<script src="/path/to/sw-toolbox/companion.js" data-service-worker="my-service-worker.js"></script>
```

### Add Service Worker Toolbox to your service worker script

In your service worker you just need to use `importScripts` to load Service Worker Toolbox

```javascript
importScripts('bower_components/sw-toolbox/sw-toolbox.js'); // Update path to match your own setup
```

### Use the toolbox

To understand how to use the toolbox read the [usage](https://googlechrome.github.io/sw-toolbox/docs/master/tutorial-usage) and [api](https://googlechrome.github.io/sw-toolbox/docs/master/tutorial-api) documentation.

## Support

If you’ve found an error in this library, please file an issue at: https://github.com/GoogleChrome/sw-toolbox/issues.

Patches are encouraged, and may be submitted by forking this project and submitting a pull request through GitHub.

## License

Copyright 2015 Google, Inc.

Licensed under the [Apache License, Version 2.0](LICENSE) (the "License");
you may not use this file except in compliance with the License. You may
obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
