import socket
import subprocess
import unittest
import time
import json
import signal
import BaseHTTPServer
import tempfile
import logging
import urllib2
import sys
import os
import threading
import portpicker


def setUpModule():
  return_code = subprocess.Popen(["go", "build", "cmd/proxy/main.go"]).wait()
  if return_code != 0:
    raise RuntimeError("Failed to build martian")


class CountingServiceHandler(BaseHTTPServer.BaseHTTPRequestHandler):

  def __init__(self, request, client_address, server):
    self._server = server
    BaseHTTPServer.BaseHTTPRequestHandler.__init__(self, request,
                                                   client_address, server)

  def do_GET(self):
    logging.info("Received a GET request. The server will respond %d this time",
                 self._server.count)
    self.send_response(200)
    self.send_header("Content-type", "text/html")
    self.end_headers()

    if self.path == "/":
      self.wfile.write("%d" % self._server.count)
      self._server.count += 1
    else:
      self.wfile.write("%s" % self.path)

    self.wfile.close()


class CountingServer(BaseHTTPServer.HTTPServer):

  def __init__(self, *args, **kw):
    succeeded = False
    while not succeeded:
      try:
        BaseHTTPServer.HTTPServer.__init__(self, *args, **kw)
        succeeded = True
      except Exception as e:
        logging.error(
          "Error binding the counting HTTP server, will retry in 1 second")
        time.sleep(1)

    self.count = 0


class MartianAndBackendServer(object):

  def __init__(self,
               martian_port,
               martian_api_port,
               backend_port,
               additional_flags=[]):
    self._martian_port = martian_port
    self._martian_api_port = martian_api_port
    self._backend_port = backend_port

    self._martian_proc = subprocess.Popen([
      os.path.join(os.getcwd(), "main"), "-addr", ":%d" % self._martian_port,
      "-api-addr", ":%d" % self._martian_api_port, "-v", "2"
    ] + additional_flags)
    self._http_server = CountingServer(("", self._backend_port),
                                       CountingServiceHandler)
    logging.info("Started martian on port %d, API port %d", self._martian_port,
                 self._martian_api_port)

    logging.info("Started http server on port %d", self._backend_port)
    self._server_thread = threading.Thread(
      target=lambda: self._http_server.serve_forever())
    self._server_thread.start()
    logging.info("Fetching http server until it's ready")
    while True:
      try:
        current_count = self.GetCurrentNumberDirectly()
        logging.info("Current http server count is %s", current_count)
        break
      except IOError:
        logging.info("Server is not ready yet...")
        time.sleep(1)
    logging.info("Http server ready.")

  def ShutDown(self):
    self._http_server.shutdown()
    self._http_server.socket.close()
    if self._martian_proc.poll() is None:
      os.kill(self._martian_proc.pid, signal.SIGINT)
      while self._martian_proc.poll() is None:
        time.sleep(1)

  def GetCurrentNumberDirectly(self):
    return int(
      urllib2.urlopen("http://localhost:%d" % self._backend_port).read())

  def GetNumberThroughProxy(self):
    return int(self.GetPathThroughProxy())

  def GetPathThroughProxy(self, path=""):
    proxy = urllib2.ProxyHandler({
      "http": "http://localhost:%d" % self._martian_port
    })
    opener = urllib2.build_opener(proxy)
    content = opener.open("http://localhost:%d%s" % (self._backend_port,
                                                     path)).read()
    return content

  def PostJsonConfigToMartian(self, config_dict):
    req = urllib2.Request("http://martian.proxy/configure")
    req.add_header("Content-Type", "application/json")
    proxy = urllib2.ProxyHandler({
      "http": "http://localhost:%d" % self._martian_port
    })
    opener = urllib2.build_opener(proxy)
    response = opener.open(req, json.dumps(config_dict)).read()
    logging.info("Posted config %r to martian, and got %r", config_dict,
                 response)


class TestCacheAndReplay(unittest.TestCase):

  def setUp(self):
    self._martian_port = portpicker.pick_unused_port()
    self._martian_api_port = portpicker.pick_unused_port()
    self._backend_port = portpicker.pick_unused_port()

    self._server_setup = MartianAndBackendServer(
      self._martian_port, self._martian_api_port, self._backend_port)

  def tearDown(self):
    self._server_setup.ShutDown()

  def testRecordAndReplayUsecase(self):
    self.assertEquals(1, self._server_setup.GetCurrentNumberDirectly())
    self.assertEquals(2, self._server_setup.GetNumberThroughProxy())

    self._server_setup.PostJsonConfigToMartian({
      "cache.Modifier": {
        "mode": "cache"
      }
    })
    self.assertEquals(3, self._server_setup.GetNumberThroughProxy())
    self._server_setup.PostJsonConfigToMartian({
      "cache.Modifier": {
        "mode": "replay"
      }
    })
    self.assertEquals(3, self._server_setup.GetNumberThroughProxy())
    # The request does not go to the backend
    self.assertEquals(4, self._server_setup.GetCurrentNumberDirectly())

  def testReplayNotCachedURL(self):
    self.assertEquals(1, self._server_setup.GetCurrentNumberDirectly())
    self._server_setup.PostJsonConfigToMartian({
      "cache.Modifier": {
        "mode": "cache"
      }
    })
    self.assertEquals(2, self._server_setup.GetNumberThroughProxy())

    self._server_setup.PostJsonConfigToMartian({
      "cache.Modifier": {
        "mode": "replay"
      }
    })
    self.assertEquals(2, self._server_setup.GetNumberThroughProxy())
    self.assertEquals("/not_cached",
                      self._server_setup.GetPathThroughProxy("/not_cached"))
    self.assertEquals(3, self._server_setup.GetCurrentNumberDirectly())


class TestSerialization(unittest.TestCase):

  def testSerAndDeser(self):
    martian_port = portpicker.pick_unused_port()
    martian_api_port = portpicker.pick_unused_port()
    backend_port = portpicker.pick_unused_port()

    cache_file = tempfile.NamedTemporaryFile()
    cache_file_path = cache_file.name
    cache_file.close()
    logging.info("Cache file location: %s", cache_file_path)

    server = MartianAndBackendServer(martian_port, martian_api_port,
                                     backend_port, ["-cache", cache_file_path])
    self.assertEquals(1, server.GetCurrentNumberDirectly())
    self.assertEquals(2, server.GetCurrentNumberDirectly())
    server.PostJsonConfigToMartian({"cache.Modifier": {"mode": "cache"}})
    self.assertEquals(3, server.GetNumberThroughProxy())
    server.ShutDown()

    server = MartianAndBackendServer(martian_port, martian_api_port,
                                     backend_port, ["-cache", cache_file_path])
    server.PostJsonConfigToMartian({"cache.Modifier": {"mode": "replay"}})
    self.assertEquals(3, server.GetNumberThroughProxy())
    server.ShutDown()


if __name__ == "__main__":
  logging.basicConfig(level=logging.DEBUG)
  unittest.main()
