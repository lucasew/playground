#!/usr/bin/env -S python3 -u

import sys
import json
import asyncio
import enum
from pathlib import Path
import os

sys.stderr = open('log.txt', 'a')

stdin = os.fdopen(sys.stdin.fileno(), mode='rb', buffering=0)


class LSPErrorCodes(enum.Enum):
    ParseError = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
    ServerNotInitialized = -32002
    UnknownErrorCode = -32001
    RequestFailed = -32803
    ServerCancelled = -32802
    ContentModified = -32801

handlers = {}

def lsp_handler(name=None):
    default_name = name
    def define_handler(func):
        name = default_name or func.__name__
        handlers[name] = func
        return func
    return define_handler
        
@lsp_handler()
def initialize(data):
    ret = dict(
        jsonrpc="2.0",
        serverInfo = dict(
            name="baguncinhad",
            version="69.420"
        ),
        capabilities=dict(
            codeActionProvider=True,
            hoverProvider=True,
            completionProvider=dict(),
            signatureHelpProvider=dict(
                triggerCharacters=["ç"],
            ),
            definitionProvider=True,
            typeDefinitionProvider=True,
            documentSymbolProvider=True,
            documentFormattingProvider=True,
            renameProvider=True,
        )
    )
    return ret
    print(json.dumps(ret), flush=True)

@lsp_handler("$/cancelRequest")
def cancel_request(data):
    task_id = data['id']
    task = tasks.get(task_id)
    if task is not None:
        task.cancel()

@lsp_handler("textDocument/hover")
def on_hover(data):
    return dict(value="never gonna give **you up**")

class LSPException(Exception):
    def __init__(self, message, code=LSPErrorCodes.InternalError, data=None):
        assert isinstance(code, LSPErrorCodes)
        super().__init__(message)
        self.message = message
        self.code = code
        self.data = data

tasks = {}        

async def handle_call(data):
    method = data['method']
    params = data.get('params')
    req_id = data['id']
    handler = handlers.get(method)
    try:
        if handler is None:
            raise LSPException("no such method", code=LSPErrorCodes.MethodNotFound)
        print(handler, file=sys.stderr, flush=True)
        result = handler(params)
        if isinstance(result, asyncio.Future):
            result = await result
        ret = dict(
            jsonrpc="2.0",
            id=req_id,
            result=result,
            error=None
        )
        print(json.dumps(ret), flush=True)
        print("<<<", ret, file=sys.stderr, flush=True)
    except LSPException as e:
        ret = dict(
            jsonrpc="2.0",
            id=req_id,
            result=None,
            error=dict(
                code=e.code,
                message=e.message,
                data=e.data
            )
        )
        print(json.dumps(ret), flush=True)
        print("<<<", ret, file=sys.stderr, flush=True)

def handle_message(data):
    req_id = data['id']
    tasks[req_id] = asyncio.create_task(handle_call(data))

def is_json_valid(data):
    try:
        json.loads(data)
        return True
    except json.JSONDecodeError:
        return False

async def main():
    print('started', handlers, file=sys.stderr, flush=True)
    buf = b""
    while True:
        line = None
        if not (b"\n" in buf or is_json_valid(buf.decode('utf-8'))):
            block = stdin.read(1) 
            if not block:
                if len(buf) == 0:
                    break
            buf += block
        if b'\n' in buf:
            line = buf[:buf.index(b'\n')+1].decode('utf-8')
            buf = buf[buf.index(b'\n')+1:]
            if not is_json_valid(line):
                print(line, file=sys.stderr, flush=True)
                continue
        else:
            line = buf.decode('utf-8')
            if is_json_valid(line):
                buf = b""
            else:
                continue
        if not line:
            continue
        print('line', line, file=sys.stderr, flush=True)
        data = json.loads(line)
        print('>>>', line, file=sys.stderr, flush=True)
        # print(tasks, file=sys.stderr, flush=True)
        handle_message(data)


asyncio.run(main())
print('finshed', file=sys.stderr, flush=True)

sys.stderr.close()
