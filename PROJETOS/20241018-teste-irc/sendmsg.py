#!/usr/bin/env python3

import socket
import ssl
import sys

from argparse import ArgumentDefaultsHelpFormatter, ArgumentParser

parser = ArgumentParser(description="send message to IRC")
parser.add_argument("--host", "-s", default="irc.stargazer-shark.ts.net")
parser.add_argument("--port", "-p", type=int, default=6697)
parser.add_argument("-u", "--user", default="bot")
parser.add_argument("-c", "--channel", default="#general")

args = parser.parse_args()

botnick = args.user
channel = args.channel

context = ssl.create_default_context()

def handle_irc(c):
    c.send(f"USER {botnick} {botnick} {botnick} :This is a fun bot!\n".encode('utf-8')) #user authentication
    c.send(f"NICK {botnick}\n".encode('utf-8'))                            #sets nick
    c.send("PRIVMSG nickserv :iNOOPE\r\n".encode('utf-8'))    #auth
    # c.send(f"JOIN {channel}\n".encode('utf-8'))        #join the chan
    for line in sys.stdin:
        c.send(f"PRIVMSG {channel} :{line}\r\n".encode('utf-8'))
    c.send("".encode('utf-8'))

with socket.create_connection((args.host, args.port)) as sock:
    with context.wrap_socket(sock, server_hostname=args.host) as ssock:
        handle_irc(ssock)
