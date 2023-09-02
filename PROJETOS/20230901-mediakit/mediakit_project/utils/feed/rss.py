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
        icon = None
        try:
            icon = entry['image']['href']
        except KeyError:
            pass

        article_content = ""
        try:
            article_content = entry['content'][0]['value']
        except KeyError:
            pass

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


