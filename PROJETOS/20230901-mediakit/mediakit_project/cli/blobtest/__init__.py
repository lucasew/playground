COMMAND_DESCRIPTION = "Only add one file to the blob store"

from mediakit_project.utils import load_module
from pathlib import Path


def command(subparser):
    def handler(args):
        module = load_module(Path(__file__).parent.parent.parent / "plugins" / "core__blob" / "__init__.py").ModuleClass()
        module.put_blob("Teste".encode('utf-8'), "demo")
    return handler
