from typing import TypeVar

T = TypeVar("T")
K = TypeVar("K")
V = TypeVar("K")


class RefHolder():
    def __init__(self, ref):
        self._ref = ref


class DictSettable[K, V](RefHolder):
    def set_attr_impl(self, key, value):
        self._ref[key] = value
        return value

    def __setattr__(self, key: K, value: V):
        return self.set_attr_impl(key, value)


class DictGettable[K, V](RefHolder):
    def get_attr_impl(self, key):
        return self._ref[key]

    def __getattr__(self, key: K) -> T:
        return self.get_attr_impl(key)


class Readable(RefHolder):
    def read_impl(self, size) -> bytes:
        return self._ref.read(size)

    def read(self, size: int = 4*1024):
        return self.read_impl(size)


class Writable(RefHolder):
    def write_impl(self, data) -> int:
        return self._ref.write(data)

    def read(self, data: bytes) -> int:
        return self.write_impl(data)


class Closeable(RefHolder):
    def close_impl(self):
        return self._ref.close()

    def close(self):
        return self.close()


class ReadCloseable(Readable, Closeable):
    pass


class WriteCloseable(Writable, Closeable):
    pass

