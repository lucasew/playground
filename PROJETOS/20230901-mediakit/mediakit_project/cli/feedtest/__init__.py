from mediakit_project.utils.feed.rss import extract
from mediakit_project.utils.feed import FeedRepository

COMMAND_DESCRIPTION = "Test with rss feeds"


def command(parser):
    parser.add_argument("feed_url", type=str)

    def handle(args):
        repo = FeedRepository(args.feed_url)
        extract(repo)
    return handle
