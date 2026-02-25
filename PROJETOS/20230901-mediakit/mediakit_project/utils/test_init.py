from . import ContextJSON

from tempfile import mktemp
from pathlib import Path


def test_exception_rollback():
    tempfile = Path(mktemp())  # the file actually don't exist

    success = False
    try:
        with ContextJSON(tempfile) as d:
            d['test'] = 2
            raise ValueError()
    except ValueError:
        success = True
    assert success
    assert not (tempfile.exists() or tempfile.is_file())
