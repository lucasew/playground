from mediakit_project.utils.feed import FeedRepository, fetch_all_feeds

COMMAND_DESCRIPTION = "Update all feeds"


def command(parser):
    def handle(args):
        for feed in fetch_all_feeds():
            pass
    return handle
