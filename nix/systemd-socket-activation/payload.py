import http.server
import socketserver
from argparse import ArgumentParser
from pathlib import Path
import os
import logging
import socket
from threading import Thread

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.DEBUG)

parser = ArgumentParser()
parser.add_argument("-o", "--out-dir", type=Path, dest='out_dir')
# parser.add_argument('-p', '--port', type=int, dest='port')

args = parser.parse_args()

logger.debug(f"env: {os.environ.values()}")
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

class CustomTCPServer(socketserver.TCPServer):
    def is_continue(self):
        return not self._BaseServer__shutdown_request

    def handle_timeout(self):
        Thread(target=lambda: self.shutdown()).start()
        logger.debug('trigger timeout, shutdown')

with CustomTCPServer(None, Handler, bind_and_activate=False) as httpd:
    httpd.socket = socket.socket(family=httpd.address_family, type=httpd.socket_type, fileno=int(os.getenv("LISTEN_FDS_FIRST_FD") or 3))
    httpd.socket.settimeout(3)
    while httpd.is_continue():
        logger.debug('checking if continue')
        httpd.handle_request()

logger.info("Stopping server")
(args.out_dir / "stop").write_text("ok")
exit(0)
