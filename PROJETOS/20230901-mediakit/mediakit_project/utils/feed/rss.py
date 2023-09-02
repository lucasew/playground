from mediakit_project import utils
from pprint import pprint
import feedparser
import time
import datetime
import logging

from . import FeedRepository, PostType, UnsupportedFeedURL

logger = logging.getLogger(__name__)


def timestamp_to_unix(timestamp):
    unix = datetime.datetime(1970, 1, 1)
    date_diff = datetime.datetime(*timestamp[0:7]) - unix
    return int(date_diff.total_seconds())


def get_feed_url_from_webpage(repo: FeedRepository) -> str:
    from urllib.request import urlopen, Request
    import re
    headers={
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"
    }
    url = repo.url
    feed_info = repo.feed_info

    if "rss_feed_url" in feed_info:
        return feed_info['rss_feed_url']

    with urlopen(Request(url, headers=headers)) as res:
        logger.info(_("Fetching webpage to get feed: {url}").format(url=url))
        first_chunk = res.read().decode('utf-8')
        matches = re.findall(r"type=\"application\/rss\+xml[^>]*href=\"([^\"]*)", first_chunk, re.MULTILINE)
        if len(matches) > 0:
            return matches[0]
    raise UnsupportedFeedURL(url)


def extract(repo: FeedRepository):
    url = repo.url

    rss_feed_url = get_feed_url_from_webpage(repo)

    logger.info(_("fetching '{url}'").format(url=url))
    data = feedparser.parse(rss_feed_url)

    head_node = data['feed']
    repo.update_meta(
        title=head_node.get('title'),
        subtitle=head_node.get('subtitle'),
        icon=head_node.get('icon'),
        rss_feed_url=rss_feed_url
    )
    for entry in data['entries']:
        published_time = entry['published_parsed']
        published_time_unix = timestamp_to_unix(published_time)

        audio_url = None
        for enclosure in entry['enclosures']:
            if enclosure.type.startswith("audio"):
                audio_url = enclosure['href']
        icon = None
        try:
            icon = entry['image']['href']
        except KeyError:
            pass

        if icon is None:
            try:
                icon = entry['media_thumbnail'][0]['url']
            except KeyError:
                pass

        article_content = ""
        try:
            article_content = entry['content'][0]['value']
        except KeyError:
            pass

        if len(article_content) == 0 and entry.get('summary') is not None:
            article_content = entry['summary']

        repo.update_post(
            entry['link'],
            published_time_unix,
            article_content=article_content,
            audio_url=audio_url,
            author=entry.get('author'),
            summary=entry['summary'],
            title=entry.get('title'),
            icon=icon
        )


