#!/usr/bin/env python

# Copyright 2015 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import argparse
import os

from subprocess import Popen, PIPE


def main(args):
  """ This function deploys the Martian web UI ("Olympus Mons") to Google App Engine.

  To deploy it, you need an existing App Engine project and username.

  Args:
    args (dict): The parsed command line parameters containing the
                 App Engine application name and optional user name 
  """
  Popen(["rm", "app.yaml"], stderr=PIPE).communicate()
  with open("app.yaml.template", "rb") as template:
    with open("app.yaml", "wb") as result:
      for line in template:
        if line.startswith("#") or len(line.strip()) == 0:
          continue
        result.write(line.replace("[[application]]", args.application))
  command = "appcfg.py "
  if args.user is not None:
    command += "-e %s " % args.user
  command += "update ."
  os.system(command)
  Popen(["rm", "app.yaml"]).communicate()


if __name__ == "__main__":
  parser = argparse.ArgumentParser(description="Deploy Martian Proxy web UI to an App Engine application."
                                   "\nMust be run from the root of the web directory.")
  parser.add_argument("-a", "--application", required=True, help="The app engine project/application to deploy.")
  parser.add_argument("-u", "--user", help="App engine username for user doing the deploy.")
  main(parser.parse_args())
