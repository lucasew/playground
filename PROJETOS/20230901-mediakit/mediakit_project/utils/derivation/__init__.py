from mediakit_project import utils

DERIVATION_HANDLERS = dict()


class DerivationNotBuiltException(Exception):
    def __init__(self, drv_hash: str, dep: str):
        super().__init__(_("Derivation is not built yet: {drv_hash} (depends on {dep_hash})").format(drv_hash=drv_hash, dep_hash=dep))


def get_derivation_dir():
    return utils.REPO_DIR / "derivations"


def get_derivation_working_dir(drv_hash: str):
    return utils.REPO_DIR / "derivations" / drv_hash


def is_derivation_built(drv_hash: str):
    return (get_derivation_working_dir(drv_hash) / "__done__").exists()


def handle_derivation(name: str, **args):
    return DERIVATION_HANDLERS[name]['fn'](**args)


def mkDerivation(handler: str, deps=[], **args):
    from uuid import uuid4
    drv_tempfile = get_derivation_dir() / f"{uuid4()}.json"
    try:
        with utils.ContextJSON(drv_tempfile) as d:
            d['handler'] = handler
            d['args'] = args
            d['deps'] = []
            for dep in deps:
                if type(dep) is str:
                    d['deps'].append(dep)
                else:
                    mkDerivation(
                        dep['handler'],
                        d['deps'],
                        **args,
                        **dep['args']
                    )
        file_hash = utils.hash_file(drv_tempfile)
        drv_dir = drv_tempfile.parent / file_hash
        drv_dir.mkdir(parents=True, exist_ok=True)
        final_file = drv_tempfile.parent / file_hash / "__meta__.json"
        drv_tempfile.rename(final_file)
        return file_hash
    finally:
        if drv_tempfile.exists():
            drv_tempfile.unlink()


def build_derivation(drv_hash: str):
    drv_dir = get_derivation_working_dir(drv_hash)
    drv_data = utils.ContextJSON(drv_dir / "__meta__.json").data
    for dep in drv_data['deps']:
        if not is_derivation_built(dep):
            raise DerivationNotBuiltException(drv_hash, dep)
    return handle_derivation(
        drv_data['handler'],
        drv_hash=drv_hash,
        **drv_data['args']
    )


def derivation_handler(name: str, pools=[], **default_config):
    with utils.repo_config as c:
        key = f"derivation:{name}"
        if key not in c:
            c[key] = {}
        for k, v in default_config:
            c[key][k] = v
        config = c[key]

    def handler(func):
        def payload(**kwargs):
            return func(config=config, **kwargs)
        DERIVATION_HANDLERS[name] = dict(
            fn=payload,
            pools=[],
        )
        return payload
    return handler


@derivation_handler("yt-dlp", pools=["download"])
def _ytdlp_video(config=None, url: str = "", download_type="bestvideo"):
    return utils.run_command_with_args("yt-dlp", url, "-f", download_type)


@derivation_handler("ffmpeg_to_h264", pools=["nvenc"], nvenc=True)
def _ffmpeg_to_h246(config=None, deps=[], drv_hash: str = ""):
    output_dir = get_derivation_working_dir(drv_hash)
    codec = "h264_nvenc" if config['nvenc'] else "h264"
    for dep in deps:
        for video in get_derivation_working_dir(dep):
            if video.name.startswith("__"):
                continue
            utils.run_command_with_args(
                "ffmpeg",
                "-vcodec", codec,
                "-i", str(video),
                f"{video.stem}.mp4",
                pwd=output_dir
            )


