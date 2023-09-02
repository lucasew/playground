from mediakit_project.utils.rss import refresh_feeds


COMMAND_DESCRIPTION = "Refresh all feeds"


def command(parser):
    def handle(args):
        for item in refresh_feeds():
            pass
    return handle
