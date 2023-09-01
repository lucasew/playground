import mediakit_project.i18n  # F401

"""CLI interface for mediakit_project project.

Be creative! do whatever you want!

- Install click or typer and create a CLI app
- Use builtin argparse
- Start a web application
- Import things from your .base module
"""

import logging
import sys
from argparse import ArgumentDefaultsHelpFormatter, ArgumentParser
from pathlib import Path
import os

from mediakit_project.utils import load_module
from mediakit_project.utils.module import ModuleClass

logger = logging.getLogger(__name__)


def add_subcommand(subparsers, name: str, submodule):
    subparser = subparsers.add_parser(name, help=submodule.COMMAND_DESCRIPTION)
    common_flags(subparser)
    handler = submodule.command(subparser)
    subparser.set_defaults(fn=handler)


DEFAULT_MEDIAKTI_PROJECT_REPO = os.environ.get("MEDIAKIT_PROJECT_REPO")


def common_flags(parser):
    parser.add_argument(
        "--repo",
        dest="repo_path",
        default=Path(DEFAULT_MEDIAKTI_PROJECT_REPO) if DEFAULT_MEDIAKTI_PROJECT_REPO is not None else None,
        type=Path,
        help=_("Where is the state repository stored")
    )
    parser.add_argument(
        "-v",
        "--verbose",
        dest="verbose",
        action="store_true",
        help=_("Give more details about what is happening"),
    )  # noqa: E501
    parser.add_argument(
        "-V",
        "--version",
        dest="is_show_version",
        action="store_true",
        help=_("Print version and exit"),
    )  # noqa: E501


def main():  # pragma: no cover
    """
    The main function executes on commands:
    `python -m mediakit_project` and `$ mediakit_project`.

    This is your program's entry point.

    You can change this function to do whatever you want.
    Examples:
        * Run a test suite
        * Run a server
        * Do some other stuff
        * Run a command line application (Click, Typer, ArgParse)
        * List all available tasks
        * Run an application (Flask, FastAPI, Django, etc.)
    """
    logging.basicConfig()
    parser = ArgumentParser(
        prog="mediakit_project", formatter_class=ArgumentDefaultsHelpFormatter
    )
    common_flags(parser)
    subparsers = parser.add_subparsers()

    for module in Path(__file__).parent.glob("*/__init__.py"):
        if str(module).find("pycache") > 0:
            continue
        module_name = module.parent.name
        subcommand_module = load_module(
            module, module_name=f"mediakit_project.cli.{module_name}"
        )
        add_subcommand(subparsers, module_name, subcommand_module)

    args = parser.parse_args()

    if args.verbose:
        logging.root.setLevel(logging.DEBUG)

    version = open(str(Path(__file__).parent.parent / "VERSION"), "r").read()
    if args.is_show_version:
        print(version)
        exit(0)
    logger.debug(f"{_('Starting')} mediakit_project v{version}")

    fn = args.__dict__.get("fn")
    args.__dict__["fn"] = None

    assert args.repo_path is not None and args.repo_path.exists() and args.repo_path.is_dir(), "State repository not specified or not a existent directory"
    ModuleClass.repo_dir = args.repo_path

    if fn is not None:
        fn(args)
    else:
        parser.parse_args([*sys.argv[1:], "--help"])
