import subprocess
import unittest
import time
import json
import BaseHTTPServer
import logging
import urllib2
import sys
import os
import threading
import portpicker


def setUpModule():
  return_code = subprocess.Popen(["go",
                                  "build",
                                  "cmd/proxy/main.go"]).wait()
  if return_code != 0:
    raise RuntimeError("Failed to build martian")


class CountingServiceHandler(BaseHTTPServer.BaseHTTPRequestHandler):

  def __init__(self, request, client_address, server):
    self._server = server
    BaseHTTPServer.BaseHTTPRequestHandler.__init__(
      self, request, client_address, server)

  def do_GET(self):
    logging.info(
      "Received a GET request. The server will respond %d this time",
      self._server.count)
    self.send_response(200)
    self.send_header('Content-type', 'text/html')
    self.end_headers()
    self.wfile.write('%d' % self._server.count)
    self.wfile.close()
    self._server.count += 1


class CountingServer(BaseHTTPServer.HTTPServer):

  def __init__(self, *args, **kw):
    BaseHTTPServer.HTTPServer.__init__(self, *args, **kw)
    self.count = 0


class TestCacheAndReplay(unittest.TestCase):

  def setUp(self):
    self._martian_port = 8889
    self._martian_api_port = 8890
    self._backend_port = 8891

    self._martian_proc = subprocess.Popen([os.path.join(os.getcwd(), "main"), "-addr", ":%d" %
                                           self._martian_port, "-api-addr", ":%d" %
                                           self._martian_api_port, "-v", "2"])
    self._http_server = CountingServer(
      ('', self._backend_port), CountingServiceHandler)
    logging.info("Started martian on port %d, API port %d",
                 self._martian_port, self._martian_api_port)

    logging.info("Started http server on port %d", self._backend_port)
    self._server_thread = threading.Thread(
      target=lambda: self._http_server.serve_forever())
    self._server_thread.start()
    logging.info("Fetching http server until it's ready")
    while True:
      try:
        current_count = self._GetCurrentNumberDirectly()
        logging.info("Current http server count is %s", current_count)
        break
      except IOError:
        logging.info("Server is not ready it...")
        time.sleep(1)
    logging.info("Http server ready.")

  def tearDown(self):
    self._http_server.shutdown()
    if self._martian_proc.poll() is None:
      self._martian_proc.terminate()
      while self._martian_proc.poll() is None:
        time.sleep(1)

  def _GetCurrentNumberDirectly(self):
    return int(urllib2.urlopen(
      "http://localhost:%d" %
      self._backend_port).read())

  def _GetNumberThroughProxy(self):
    proxy = urllib2.ProxyHandler(
      {'http': 'http://localhost:%d' % self._martian_port})
    opener = urllib2.build_opener(proxy)
    content = opener.open("http://localhost:%d" % self._backend_port).read()
    return int(content)

  def _PostJsonConfigToMartian(self, config_dict):
    req = urllib2.Request("http://martian.proxy/configure")
    req.add_header('Content-Type', 'application/json')
    proxy = urllib2.ProxyHandler(
      {'http': 'http://localhost:%d' % self._martian_port})
    opener = urllib2.build_opener(proxy)
    response = opener.open(req, json.dumps(config_dict)).read()
    logging.info(
      "Posted config %r to martian, and got %r",
      config_dict,
      response)

  def testRecordAndReplayUsecase(self):
    self.assertEquals(1, self._GetCurrentNumberDirectly())
    self.assertEquals(2, self._GetNumberThroughProxy())

    self._PostJsonConfigToMartian({'cache.Modifier': {'mode': "cache"}})
    self.assertEquals(3, self._GetNumberThroughProxy())
    self._PostJsonConfigToMartian({'cache.Modifier': {'mode': "replay"}})
    self.assertEquals(3, self._GetNumberThroughProxy())
    self.assertEquals(4, self._GetCurrentNumberDirectly())

if __name__ == "__main__":
  logging.basicConfig(level=logging.DEBUG)
  unittest.main()
