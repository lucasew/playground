import dagger
from dagger import dag, function, object_type


@object_type
class Playground:
    @function
    def container_echo(self, string_arg: str) -> dagger.Container:
        """Returns a container that echoes whatever string argument is provided"""
        return dag.container().from_("alpine:latest").with_exec(["echo", string_arg])

    @function
    async def grep_dir(self, directory_arg: dagger.Directory, pattern: str) -> str:
        """Returns lines that match a pattern in the files of the provided Directory"""
        return await (
            dag.container()
            .from_("alpine:latest")
            .with_mounted_directory("/mnt", directory_arg)
            .with_workdir("/mnt")
            .with_exec(["grep", "-R", pattern, "."])
            .stdout()
        )

    @function
    async def alpine_with(self, packages: str) -> dagger.Container:
        return await (
            dag.container()
                .from_("alpine:latest")
                .with_exec(["apk", "add", *(packages.split(','))])
        )

    @function
    async def basic_http_server(self, content: str ="Hello, world") -> dagger.Container:
        return (
        dag.container()
            .from_("nginx")
            .with_new_file("/usr/share/nginx/html/index.html", content)
            .with_exposed_port(80)
            .as_service()
        )

