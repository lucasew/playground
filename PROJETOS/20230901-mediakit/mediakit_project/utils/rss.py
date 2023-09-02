from mediakit_project import utils
from pprint import pprint
import feedparser
import time
import datetime
import logging

logger = logging.getLogger(__name__)


def get_root_dir():
    return utils.REPO_DIR / "rss"


def get_feed_dir(url: str):
    return get_root_dir() / utils.hash_string(url)


def timestamp_to_unix(timestamp):
    unix = datetime.datetime(1970, 1, 1)
    date_diff = datetime.datetime(*timestamp[0:7]) - unix
    return int(date_diff.total_seconds())


def fetch_rss(url: str):
    logger.info(_("fetching '{url}'").format(url=url))
    data = feedparser.parse(url)
    feed_dir = get_feed_dir(url)

    head_node = data['feed']
    head_data = dict(
        url=url,
        subtitle=head_node.get('subtitle'),
        title=head_node.get('title'),
    )
    with utils.ContextJSON(feed_dir / "__meta__.json") as d:
        d['url'] = url
        d['title'] = head_node.get('title_detail')
        d['subtitle'] = head_node.get('subtitle')
        d['image'] = head_node.get('image')
    for entry in data['entries']:
        published_time = entry['published_parsed']
        with utils.ContextJSON(feed_dir / f"{timestamp_to_unix(published_time)}.json") as d:
            for k in ["links", "summary", "title", "comments", "enclosures", "published", "expired"]:
                d[k] = entry.get(k)




