# from mediakit_project.utils.feed.rss import refresh_feeds


COMMAND_DESCRIPTION = "Refresh all feeds"


def command(parser):
    def handle(args):
        raise NotImplementedError()
        pass
        # for item in refresh_feeds():
        #     pass
    return handle
