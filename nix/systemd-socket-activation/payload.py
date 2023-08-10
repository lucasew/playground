import http.server
import socketserver
from argparse import ArgumentParser
from pathlib import Path
import os
import logging
import socket

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.DEBUG)

parser = ArgumentParser()
parser.add_argument("-o", "--out-dir", type=Path, dest='out_dir')
# parser.add_argument('-p', '--port', type=int, dest='port')

args = parser.parse_args()

# assert os.getenv("LISTEN_FDS") == "3"
logger.info("Starting")

(args.out_dir / "init").write_text("ok")

def iter_incr():
    i = 0
    while True:
        yield i
        i += 1

reqs_iter = iter_incr()


class Handler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        logger.info("Receiving request")
        (args.out_dir / f"request{next(reqs_iter)}").write_text("ok")

with socketserver.TCPServer(None, Handler, bind_and_activate=False) as httpd:
    httpd.socket = socket.socket(family=httpd.address_family, type=httpd.socket_type, fileno=3)
    httpd.serve_forever()
logger.info("Stopping server")
