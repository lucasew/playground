from mediakit_project.utils.validation import VALIDATORS

from typing import Protocol, runtime_checkable


@runtime_checkable
class Readable(Protocol):
    def read(self, size: int) -> bytes:
        return b''


@runtime_checkable
class Writable(Protocol):
    def write(self, data: bytes) -> int:
        return 0

