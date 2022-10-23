from pathlib import Path
import asyncio
from shutil import which


def generate_random_string(length: int = 8):
    from random import choice
    from string import ascii_lowercase
    ret = []
    for i in range(length):
        ret.append(choice(ascii_lowercase))
    return "".join(ret)


def quote_args(*args: str):
    from shlex import join
    return join(args)


class PyRUN:
    docker_bin = which("docker")

    def __init__(self, docker_bin=which("docker"), log_dir=None):
        PyRUN.docker_bin = docker_bin
        self.log_dir = log_dir

    class DockerArgumentGenerator():
        def __init__(self):
            self.docker_args = []
            self.app_args = []

        def __repr__(self):
            return f"DockerArgumentGenerator Docker={self.docker_args} App={self.app_args}"

    class app_args(DockerArgumentGenerator):
        def __init__(self, *args):
            super().__init__()
            self.app_args = args

    class docker_args(DockerArgumentGenerator):
        def __init__(self, *args):
            super().__init__()
            self.docker_args = args

    class flatten_args(DockerArgumentGenerator):
        def rabbit_hole(self, arg):
            _docker_args = []
            _app_args = []
            for docker_arg in arg.docker_args:
                if issubclass(type(docker_arg), PyRUN.DockerArgumentGenerator):
                    new_docker_args, new_app_args = self.rabbit_hole(docker_arg)
                    for new_docker_arg in new_docker_args:
                        _docker_args.append(new_docker_arg)
                    for new_app_arg in new_app_args:
                        _app_args.append(new_app_arg)
                else:
                    _docker_args.append(str(docker_arg))
            for app_arg in arg.app_args:
                if issubclass(type(app_arg), PyRUN.DockerArgumentGenerator):
                    new_docker_args, new_app_args = self.rabbit_hole(app_arg)
                    for new_docker_arg in new_docker_args:
                        _docker_args.append(new_docker_arg)
                    for new_app_arg in new_app_args:
                        _app_args.append(new_app_arg)
                else:
                    _app_args.append(str(app_arg))
            return _docker_args, _app_args

        def __init__(self, *args):
            super().__init__()
            for arg in args:
                new_docker_args, new_app_args = self.rabbit_hole(arg)
                for new_docker_arg in new_docker_args:
                    self.docker_args.append(new_docker_arg)
                for new_app_arg in new_app_args:
                    self.app_args.append(new_app_arg)

    class local_path(DockerArgumentGenerator):
        def __init__(self, path: Path):
            super().__init__()
            mountpoint = generate_random_string()
            self.docker_args = ["-v", f"{str(path)}:/{mountpoint}"]
            self.app_args = [f"/{mountpoint}"]

    class local_port(DockerArgumentGenerator):
        def __init__(self, port=None, local_port=None, host_port=None, expand_port=True):
            super().__init__()
            internal_port = local_port or port
            external_port = host_port or port
            assert internal_port is not None and external_port is not None
            self.docker_args = ["-p" f"{internal_port}:{external_port}"]
            if external_port:
                self.app_args = [str(internal_port)]

    class text_file(DockerArgumentGenerator):
        def __init__(self, content: str, encoding=None, executable=False, writable=False):
            super().__init__()
            from os import chmod
            import stat
            perm = 0
            perm |= stat.S_IREAD | stat.S_IRGRP | stat.S_IROTH #  read
            if executable:
                perm |= stat.S_IEXEC | stat.S_IXGRP | stat.S_IXOTH
            if writable:
                perm |= stat.S_IWRITE | stat.S_IWGRP | stat.S_IWOTH
            from tempfile import NamedTemporaryFile
            file = NamedTemporaryFile(mode='w', delete=False)
            file.write(content)
            file.flush()
            file_path = Path(file.name)
            chmod(str(file_path), perm)
            volume = PyRUN.local_path(file_path)
            self.docker_args = volume.docker_args
            self.app_args = volume.app_args

    class escape_args(DockerArgumentGenerator):
        def __init__(self, *args):
            super().__init__()
            expanded = PyRUN.flatten_args(*args)
            print(expanded)
            self.docker_args = expanded.docker_args
            self.app_args = [quote_args(*expanded.app_args)]
            print(expanded.app_args)

    def run_script(script="", interpreter='sh'):
        return PyRUN.app_args(PyRUN.text_file(f"""#!/usr/bin/env {interpreter}
    {script}
        """, executable=True))

    async def async_run(*args, wait=False, **kwargs):
        print("Running", args)
        proc = await asyncio.create_subprocess_exec(*args, **kwargs)
        if wait:
            await proc.wait()
        return proc

    class DockerCommand:
        def __init__(self, *args, image: str = "alpine", enable_tty=True, enable_interactive=True, **kwargs):
            args_to_process = []
            args_to_process.append(PyRUN.docker_args("run"))
            if enable_tty:
                args_to_process.append(PyRUN.docker_args("-t"))
            if enable_interactive:
                args_to_process.append(PyRUN.docker_args("-i"))
            args_to_process.append(PyRUN.docker_args("-d"))
            for arg in args:
                args_to_process.append(arg)
            args_to_process.append(PyRUN.docker_args(image))
            ret = PyRUN.flatten_args(*args_to_process)
            ret = [PyRUN.docker_bin, *ret.docker_args, *ret.app_args]
            self.command = ret
            print(quote_args(*ret))
            print(ret)
            self.running = False
            # return ret

        async def run(self):
            if self.running:
                return
            proc = await PyRUN.async_run(*self.command, stdout=asyncio.subprocess.PIPE, wait=True)
            container_id = await proc.stdout.readline()
            self.container_id = container_id.decode('ascii').strip()
            self.running = True

        async def stop(self):
            if not self.running:
                return
            await PyRUN.async_run(PyRUN.docker_bin, "stop", self.container_id, wait=True)

        async def wait(self):
            if not self.running:
                return
            print(self.container_id)
            await PyRUN.async_run(PyRUN.docker_bin, "wait", self.container_id, wait=True)

    async def sequential(*cmds):
        for cmd in cmds:
            await cmd.run()
            await cmd.wait()

    async def concurrent(*cmds):
        await asyncio.gather(*[cmd.run() for cmd in cmds])
        await asyncio.gather(*[cmd.wait() for cmd in cmds])

async def main():
    await PyRUN.concurrent(
        PyRUN.DockerCommand(PyRUN.app_args("sh", "-c"), PyRUN.escape_args(PyRUN.app_args("cat"), PyRUN.text_file(content="Hello, world")), image='alpine'),
        PyRUN.DockerCommand(PyRUN.run_script("""
echo $HOSTNAME
apk add neofetch
neofetch
    """), image='alpine')
    )

asyncio.run(main())

