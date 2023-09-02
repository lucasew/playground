import logging
import json
from pathlib import Path
import configparser
from uuid import uuid4 as uuidgen

logger = logging.getLogger(__name__)

REPO_DIR = None


def load_module(script_path, module_name="module"):
    import importlib
    logger.debug(
        _("Loading module '{module_path}' ...").format(module_path=script_path)
    )
    spec = importlib.util.spec_from_file_location(module_name, script_path)
    assert spec is not None, _(
        "Can't import module at '{module_path}'"
    ).format(module_path=script_path)
    model_script = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(model_script)

    return model_script


def hash_string(text: str) -> str:
    from hashlib import sha256
    hasher = sha256()
    hasher.update(text.encode('utf-8'))
    return hasher.hexdigest()


def hash_file(file: Path) -> str:
    from hashlib import sha256
    hasher = sha256()
    with file.open('rb') as f:
        while True:
            buf = f.read(4096)
            if not buf:
                break
            hasher.update(buf)
    return hasher.hexdigest()


class ContextConfig():
    def __init__(self, configfile="mediakit_project.conf", readonly=True):
        self.configfile = REPO_DIR / configfile
        self.mode = "r" if readonly else "w"

    def __enter__(self):

        config = configparser.ConfigParser()
        config.read(str(self.configfile))
        self._config = config
        return config

    def __exit__(self):
        with open(self._configfile, self.mode) as f:
            self._config.write(f)


class ContextJSON():
    def __init__(self, file):
        self.file = file

    def __enter__(self):
        if not self.file.parent.exists():
            self.file.parent.mkdir(exist_ok=True, parents=True)
        if not self.file.exists():
            self.file.write_text("{}")
        with open(str(self.file), 'r') as f:
            self._data = json.load(f)
        return self._data

    def __exit__(self, exc_type, exc_val, exc_traceback):
        tmpfile = self.file.parent / f".{uuidgen()}.json"
        try:
            with tmpfile.open("w") as f:
                json.dump(self._data, f, indent=4, sort_keys=True)
            tmpfile.rename(self.file)
        finally:
            if tmpfile.exists():
                tmpfile.unlink()

