from mediakit_project.utils.rss import fetch_rss

COMMAND_DESCRIPTION = "Test with rss feeds"


def command(parser):
    parser.add_argument("feed_url", type=str)

    def handle(args):
        fetch_rss(args.feed_url)
    return handle
