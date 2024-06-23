#!/usr/bin/env -S python3 -u

import sys
import json
import asyncio
import enum
from pathlib import Path
import os
import threading
import functools

sys.stderr = open('log.txt', 'a')

log = functools.partial(print, file=sys.stderr, flush=True)

stdin = os.fdopen(sys.stdin.fileno(), buffering=0, mode='rb')
stdout = os.fdopen(sys.stdout.fileno(), buffering=0, mode='wb')

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
            executeCommandProvider=dict(
              commands=[]  
            ),
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

@lsp_handler("textDocument/codeAction")
def code_action(data):
    select_range = data['range']
    is_same_line = select_range['start']['line'] == select_range['end']['line'] and (select_range['end']['character'] - select_range['start']['character']) == 1
    is_next_line = (select_range['end']['line'] - select_range['start']['line']) == 1 and select_range['end']['character'] == 0
    is_one_char_selection = is_same_line or is_next_line
    return [
        dict(
            title="Print current uid",
            kind="refactor",
            edit=dict(
                changes={
                    data['textDocument']['uri']: [
                        dict(
                             range=dict(
                                 start=select_range['start'],
                                 end=select_range['start'] if is_one_char_selection else select_range['end']
                             ),
                             newText=str(os.getuid())
                         )
                    ]
                }
            ),
        )
    ]

# @lsp_handler("workspace/executeCommand")
# def execute_command(data):
#     command = data['command']
#     arguments = data['arguments']
#     if command == "whoami" and arguments[0].startswith("file:///"):
#         file = arguments[0]
#         file = file.replace('file://', '')
#         log('file', file)
#         with open(file, 'a') as f:
#             print(os.getuid(), file=f)
#     raise LSPException("no such command", code=LSPErrorCodes.MethodNotFound)
    

@lsp_handler("textDocument/hover")
def on_hover(data):
    return dict(contents="# hover\n\nnever gonna give **you up**")


class LSPException(Exception):
    def __init__(self, message, code=LSPErrorCodes.InternalError, data=None):
        assert isinstance(code, LSPErrorCodes)
        super().__init__(message)
        self.message = message
        self.code = code
        self.data = data

tasks = {}        

async def handle_call(data):
    log(data)
    method = data['method']
    params = data.get('params')
    req_id = data.get('id')
    handler = handlers.get(method)
    log("call", req_id, method, params)
    try:
        if handler is None:
            raise LSPException("no such method", code=LSPErrorCodes.MethodNotFound)
        log(handler)
        result = handler(params)
        if isinstance(result, asyncio.Future):
            result = await result
        ret = dict(
            jsonrpc="2.0",
            id=req_id,
            result=result,
            error=None
        )
        log(">>>", ret)
        writer.write(ret)
        # sys.stdout.write(json.dumps(ret).encode('utf-8'))
        
    except LSPException as e:
        ret = dict(
            jsonrpc="2.0",
            id=req_id,
            result=None,
            error=dict(
                code=e.code.value,
                message=e.message,
                data=e.data
            )
        )
        log(">>>", ret)
        writer.write(ret)
        # sys.stdout.write(json.dumps(ret).encode('utf-8'))

# https://github.com/python-lsp/python-lsp-jsonrpc/blob/786d8dd8f830dbd83a17962c0167183a6609e72f/pylsp_jsonrpc/streams.py
class JsonRpcStreamReader:
    def __init__(self, rfile):
        self._rfile = rfile

    def close(self):
        self._rfile.close()

    def _read_message(self):
        """Reads the contents of a message.

        Returns:
            body of message if parsable else None
        """
        line = self._rfile.readline()

        if not line:
            return None

        content_length = self._content_length(line)

        # Blindly consume all header lines
        while line and line.strip():
            line = self._rfile.readline()

        if not line:
            return None

        # Grab the body
        return self._rfile.read(content_length)

    @staticmethod
    def _content_length(line):
        """Extract the content length from an input line."""
        if line.startswith(b'Content-Length: '):
            _, value = line.split(b'Content-Length: ')
            value = value.strip()
            try:
                return int(value)
            except ValueError as e:
                raise ValueError(f"Invalid Content-Length header: {value}") from e

        return None

class JsonRpcStreamWriter:
    def __init__(self, wfile, **json_dumps_args):
        self._wfile = wfile
        self._wfile_lock = threading.Lock()
        self._json_dumps_args = json_dumps_args

    def close(self):
        with self._wfile_lock:
            self._wfile.close()

    def write(self, message):
        with self._wfile_lock:
            if self._wfile.closed:
                return
            try:
                body = json.dumps(message, **self._json_dumps_args)

                # Ensure we get the byte length, not the character length
                content_length = len(body) if isinstance(body, bytes) else len(body.encode('utf-8'))

                response = (
                    f"Content-Length: {content_length}\r\n"
                    f"Content-Type: application/vscode-jsonrpc; charset=utf8\r\n\r\n"
                    f"{body}"
                )

                self._wfile.write(response.encode('utf-8'))
                self._wfile.flush()
            except Exception:  # pylint: disable=broad-except
                log("Failed to write message to output file %s", message)

async def handle_message(data):
    await handle_call(data)
    # req_id = data['id']
    # tasks[req_id] = asyncio.create_task(handle_call(data))

def is_json_valid(data):
    try:
        json.loads(data)
        return True
    except json.JSONDecodeError:
        return False

reader = JsonRpcStreamReader(stdin)
writer = JsonRpcStreamWriter(stdout)

async def main():
    log('started', handlers)
    while True:
        message = reader._read_message()
        if message is None:
            continue
        log("got here")
        message = json.loads(message)
        log("<<<", message)
        await handle_message(message)

asyncio.run(main())
log('finshed')

sys.stderr.close()
