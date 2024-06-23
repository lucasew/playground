#!/usr/bin/env -S python3 -u

import sys
import json
import asyncio
import enum
from pathlib import Path
import os

sys.stderr = open('log.txt', 'a')

stdin_buf_queue = asyncio.Queue()

running = True

async def read_stdin_hook(buf):
    print('buf', buf, file=sys.stderr, flush=True)
    if not buf:
        return
    stdin_buf_queue.put(buf)


stdin_lines_queue = asyncio.Queue()
async def read_stdin_queue_lines():
    buf = b""
    while True:
        if not stdin_buf_queue.empty():
            block = stdin_buf_queue.get_nowait()
            if block is not None:
                buf += block
        if b'\n' in buf:
            nl_index = buf.index(b'\n')
            line = buf[:nl_index].decode('utf-8')
            buf = buf[nl_index:]
            stdin_lines_queue.put(buf)
            continue
        if is_json_valid(buf.decode('utf-8')):
            line = buf.decode('utf-8')
            buf = b""
            stdin_lines_queue.put(line)
            print("<<<", line, file=sys.stderr, flush=True)
            continue


stdout_line_queue = asyncio.Queue()
async def write_stdout_queue():
    while True:
        try:
            item = stdout_line_queue.get_nowait()
            if not isinstance(item, str):
                item = json.dumps(item)
            print(">>>", item, file=sys.stderr, flush=True)
            print(item, flush=True)
        except asyncio.QueueEmpty:
            if running:
                await asyncio.sleep(0.1)
            else:
                break


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
        stdout_line_queue.put(dict(
            jsonrpc="2.0",
            id=req_id,
            result=result,
            error=None
        ))
    except LSPException as e:
        stdout_line_queue.put(dict(
            jsonrpc="2.0",
            id=req_id,
            result=None,
            error=dict(
                code=e.code,
                message=e.message,
                data=e.data
            )
        ))

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
    asyncio.get_event_loop().add_reader(sys.stdin.fileno(), read_stdin_hook)
    read_stdin_lines_task = asyncio.create_task(read_stdin_queue_lines())
    write_stdout_task = asyncio.create_task(write_stdout_queue())
    print('started', handlers, file=sys.stderr, flush=True)
    while True:
        try:
            line = stdin_lines_queue.get_nowait()
            data = json.loads(line)
            handle_message(data)
        except json.JSONDecodeError as e:
            print(e, file=sys.stderr, flush=True)
        except asyncio.QueueEmpty:
            await asyncio.sleep(0.1)

asyncio.run(main())
print('finshed', file=sys.stderr, flush=True)

sys.stderr.close()
