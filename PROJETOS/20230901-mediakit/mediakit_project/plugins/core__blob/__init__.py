from mediakit_project.utils import module
from mediakit_project.utils.type_hints import Readable, Writable

from typing import Optional, Any
from hashlib import sha256
from uuid import uuid4
import io
from pathlib import Path


class BlobNotFoundException(Exception):
    def __init__(self, blob_id: str, variation: Optional[str] = None):
        variation_suffix = f"@{variation}" if variation is not None else ""
        super(_("Blob '{blob}' not found").format(blob=blob_id + variation_suffix))


class BlobInvalidDataToPutInABlob(Exception):
    def __init__(self, stream):
        dtype = type(stream)
        super(_("Invalid data to put in a blob: {dtype}").format(dtype=f"{dtype.__module__}.{dtype.__name__}"))


class ModuleClass(module.ModuleClass):
    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def _get_blob_repo_dir(self, variation: Optional[str] = None):
        blob_dir = self.repo_dir / "blob"
        if variation is not None:
            blob_dir = blob_dir / variation
        blob_dir.mkdir(parents=True, exist_ok=True)
        return blob_dir

    def _get_blob_file(self, blob_id: str, variation: Optional[str] = None):
        blob_repo_dir = self._get_blob_repo_dir(variation=variation)
        if variation is not None:
            blob_file = blob_repo_dir / blob_id
            return blob_file if blob_file.exists() else None
        for blob_dir in self._get_blob_repo_dir().iterdir():
            if not blob_dir.is_dir():
                continue
            blob_file = blob_dir / blob_id
            if blob_file.exists():
                return blob_file
        return None

    def get_blob(self, blob_id: str, variation: Optional[str] = None) -> Readable:
        blob_file_path = self._get_blob_file(blob_id, variation=variation)
        if blob_file_path is None:
            raise BlobNotFoundException(blob_id, variation)
        return open(str(blob_file_path), 'rb')

    def put_blob_stream(self, stream: Writable, variation="default") -> str:
        temp_file = self._get_blob_repo_dir("__cache__") / str(uuid4())
        hasher = sha256()
        with temp_file.open('wb') as f:
            while True:
                buf = stream.read(16*1024)
                if not buf:
                    break
                hasher.update(buf)
                f.write(buf)
        blob_id = hasher.hexdigest()
        output_file = self._get_blob_repo_dir(variation=variation) / blob_id
        temp_file.rename(output_file)
        if temp_file.exists():
            temp_file.unlink()
        return blob_id

    def put_blob_file(self, file: Path, variation="default") -> str:
        with file.open('rb') as f:
            return self.put_blob_stream(f, variation=variation)

    def put_blob_bytes(self, data: bytes, variation="default") -> str:
        return self.put_blob_stream(io.BytesIO(data), variation=variation)

    def put_blob(self, data: Any, variation="default") -> str:
        if isinstance(data, Path):
            return self.put_blob_file(data, variation=variation)
        elif isinstance(data, bytes):
            return self.put_blob_bytes(data, variation=variation)
        elif isinstance(data, Readable):
            return self.put_blob_stream(data, variation=variation)
        else:
            raise BlobInvalidDataToPutInABlob(data)

