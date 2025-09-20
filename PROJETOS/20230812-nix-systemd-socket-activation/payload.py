import http.server
import socketserver
from argparse import ArgumentParser
from pathlib import Path
import os
import logging
import socket
from threading import Thread
import sys

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.DEBUG)

parser = ArgumentParser()
parser.add_argument("-o", "--out-dir", type=Path, dest='out_dir')
parser.add_argument("-u", "--unlock-convenience-limits", type=Path, dest='unlock_convenience_limits')
# parser.add_argument('-p', '--port', type=int, dest='port')

args = parser.parse_args()

logger.debug(f"env: {os.environ.values()}")

assert int(os.getenv("LISTEN_FDS")) > 0

logger.info("Starting")

(args.out_dir / "init").write_text("ok")

def iter_incr():
    i = 0
    while True:
        yield i
        i += 1

def limit_calls(amount: int):
    call_number = iter_incr()
    def __ret(func):
        if next(call_number) < amount or args.unlock_convenience_limits:
            return func()
    return __ret
logger_limit = limit_calls(10)

reqs_iter = iter_incr()

threads = []



class Handler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        logger.info("Receiving request")
        (args.out_dir / f"request{next(reqs_iter)}").write_text("ok")

class CustomTCPServer(socketserver.TCPServer):
    def is_continue(self):
        logger_limit(lambda: logger.debug(f"Continue {not self._BaseServer__shutdown_request}"))
        return not self._BaseServer__shutdown_request

    def handle_timeout(self):
        self._BaseServer__shutdown_request = True
        logger.debug('trigger timeout, shutdown')

httpd = CustomTCPServer(None, Handler, bind_and_activate=False)
# httpd.socket = socket.socket(family=httpd.address_family, type=httpd.socket_type, fileno=int(os.getenv("LISTEN_FDS_FIRST_FD") or 3))
httpd.socket = socket.fromfd(int(os.getenv("LISTEN_FDS_FIRST_FD") or 3), family=httpd.address_family, type=httpd.socket_type)
# httpd.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 4)
httpd.socket.settimeout(3)
while httpd.is_continue():
    logger_limit(lambda: logger.debug('checking if continue'))
    httpd.handle_request()
# httpd.socket.close()

(args.out_dir / "stop").write_text("ok")

logger.info("Stopping server")

# for thread in threads:
#     thread.join()
# sys.exit(0)
exit(0)
# os.kill(os.getpid(), 9)
