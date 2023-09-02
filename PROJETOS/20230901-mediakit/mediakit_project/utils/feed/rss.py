from mediakit_project import utils
from pprint import pprint
import feedparser
import time
import datetime
import logging

from . import FeedRepository, PostType

logger = logging.getLogger(__name__)


def timestamp_to_unix(timestamp):
    unix = datetime.datetime(1970, 1, 1)
    date_diff = datetime.datetime(*timestamp[0:7]) - unix
    return int(date_diff.total_seconds())


def extract(repo: FeedRepository):
    url = repo.url

    logger.info(_("fetching '{url}'").format(url=url))
    data = feedparser.parse(url)
    pprint(data)

    head_node = data['feed']
    repo.update_meta(
        title=head_node.get('title'),
        subtitle=head_node.get('subtitle'),
        icon=head_node.get('icon')
    )
    for entry in data['entries']:
        published_time = entry['published_parsed']
        published_time_unix = timestamp_to_unix(published_time)

        audio_url = None
        for enclosure in entry['enclosures']:
            if enclosure.type.startswith("audio"):
                audio_url = enclosure['href']

        repo.update_post(
            entry['link'],
            published_time_unix,
            article_content=entry.get('content'),
            audio_url=audio_url,
            author=entry.get('author'),
            summary=entry['summary'],
            title=entry.get('title'),
        )






def refresh_feeds():
    feed_dir = get_root_dir()
    feeds = []
    for feed in feed_dir.iterdir():
        if not feed.is_dir():
            continue
        if feed.name.startswith("__"):
            continue
        with utils.ContextJSON(feed / "__meta__.json") as d:
            last_updated = d['_last_updated']
            url = d['url']
            feeds.append(dict(last_updated=last_updated, url=url))
    feeds.sort(key=lambda x: x['last_updated'])
    for feed in feeds:
        yield fetch_rss(feed['url'])


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
        d['_last_updated'] = int(time.time())

    for entry in data['entries']:
        published_time = entry['published_parsed']
        with utils.ContextJSON(feed_dir / f"{timestamp_to_unix(published_time)}.json") as d:
            for k in ["links", "summary", "title", "comments", "enclosures", "published", "expired"]:
                d[k] = entry.get(k)




