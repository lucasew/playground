#!/usr/bin/env -S sd nix shell 
#!nix-shell -i python3 -p python3Packages.systemd

from systemd import journal

import socket
import ssl
import sys
import asyncio
from collections import defaultdict
import threading
from concurrent.futures import ThreadPoolExecutor

from argparse import ArgumentDefaultsHelpFormatter, ArgumentParser

# doesn't work well when there is a spammy service

parser = ArgumentParser(description="send message to IRC")
parser.add_argument("--host", "-s", default="irc.stargazer-shark.ts.net")
parser.add_argument("--port", "-p", type=int, default=6697)

args = parser.parse_args()

prefix = f"j-{socket.gethostname()}"

botnick = prefix

context = ssl.create_default_context()

loop = asyncio.get_event_loop()

buf = defaultdict(lambda: [])

buf['general'].append('started')

journal_queue = asyncio.Queue()

pool = ThreadPoolExecutor()

def journal_source():
    print("!init", "journal_source")
    j = journal.Reader()
    j.seek_tail()
    j.get_previous()
    while True:
        for entry in j:
            print("!journal", entry)
            chan = f"{prefix}.{entry.get('_SYSTEMD_UNIT', 'stderr')}".replace('.', '__')
            message = entry['MESSAGE']
            buf[chan].append(message)
            print(chan, message)

async def handle_irc(w):
    w.write(f"USER {botnick} {botnick} {botnick} :This is a fun bot!\n".encode('utf-8')) #user authentication
    w.write(f"NICK {botnick}\n".encode('utf-8'))                            #sets nick
    w.write("PRIVMSG nickserv :iNOOPE\r\n".encode('utf-8'))    #auth
    await w.drain()

    joined = set()
    while True:
        await asyncio.sleep(1.0)
        if w.is_closing():
            exit(1)
        to_delete = []
        for chan in buf.keys():
            print('!flush', chan)
            if len(buf[chan]) == 0:
                continue
            # if chan not in joined:
            w.write(f"JOIN #{chan}\n".encode('utf-8'))
                # joined.add(chan)
            message = buf[chan][-1]
            if len(buf[chan]) > 1:
                w.write(f"PRIVMSG #{chan}  :[INFO] Skipping {len(buf[chan])} messages because of flood\r\n".encode('utf-8'))
            w.write(f"PRIVMSG #{chan}  :{message}\r\n".encode('utf-8'))
            await w.drain()
            to_delete.append(chan)
        for chan in to_delete:
            buf.pop(chan)

async def read_printer(r):
    while True:
        line = (await r.readuntil('\n'.encode('utf-8'))).decode('utf-8').strip()
        if r.at_eof():
            exit(1)
        print('!server', line)

async def main():
    r, w = await asyncio.open_connection(args.host, args.port, ssl=context)

    await asyncio.wait([
        asyncio.create_task(handle_irc(w)),
        asyncio.create_task(read_printer(r)),
        pool.submit(journal_source)
    ], return_when=asyncio.FIRST_COMPLETED)
    exit(0)

asyncio.run(main())
