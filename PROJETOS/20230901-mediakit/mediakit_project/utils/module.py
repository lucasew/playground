
class ModuleClass:
    repo_dir = None

    def __init__(self, **kwargs):
        assert ModuleClass.repo_dir is not None
        self.repo_dir = ModuleClass.repo_dir
