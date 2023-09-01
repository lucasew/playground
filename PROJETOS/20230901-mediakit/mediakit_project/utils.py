import logging

logger = logging.getLogger(__name__)


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


class ModuleClass:
    repo_dir = None

    def __init__(self, **kwargs):
        assert ModuleClass.repo_dir is not None
        self.repo_dir = ModuleClass.repo_dir
