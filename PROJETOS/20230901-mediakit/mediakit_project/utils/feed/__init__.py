from mediakit_project import utils
from enum import Enum
from typing import Optional
import time
import logging
from pprint import pprint

__DOC__ = """
Feeds are anything that provides posts over time

Examples:
- YouTube channels
- YouTube playlists
- Twitter accounts
- RSS Feeds (basically anything else)

Each feed can provide the folloing set of resources
- Video: the content itself is a video, like a YouTube video that may have
to be reencoded.
- Audio: the content itself is a audio, like a Podcast.
- Article: the content itself is some kind of post, the shownotes of
a podcast can also be considered an article.

These resources are extracted by modules or external software
such as yt-dlp and may have to be reencoded using ffmpeg, for example.

By default, all feeds provide articles.

All extraction must be deferred to be done later. The queue is built using
a build.ninja file that can be run on another machine that is syncing the
repo.

All feed urls should be the original source URL and any third party resources
must be handled by extractors.
"""

logger = logging.getLogger(__name__)


class UnsupportedFeedURL(Exception):
    def __init__(self, url: str):
        super().__init__(_("UnsupportedFeedURL: {url}").format(url=url))


class UnsupportedPostURL(Exception):
    def __init__(self, url: str):
        super().__init__(_("UnsupportedPostURL: {url}").format(url=url))


class PostType(Enum):
    Article = 0
    Video = 0
    Audio = 0


def get_root_dir():
    return utils.REPO_DIR / "feeds"


def get_feed_dir(url: str):
    return get_root_dir() / utils.hash_string(url)


def fetch_all_feeds():
    from json import loads
    feeds = []
    for feed in get_root_dir().iterdir():
        if feed.name.startswith("__"):
            continue
        data = loads((feed / "__meta__.json").read_text())
        repository = FeedRepository(data['url'])
        feed_info = repository.feed_info
        feeds.append(dict(
            feed_info=feed_info,
            repository=repository,
            feed=feed,
        ))
    feeds.sort(key=lambda feed: feed['feed_info']['last_updated'])
    for feed in feeds:
        update_one_feed(feed['repository'])
        yield feed['repository']


def update_one_feed(feed):
    from .rss import extract  # TODO: add other filters
    extract(feed)


class FeedRepository():
    def __init__(self, url: str):
        self._url = url
        self.feed_repo = get_root_dir() / utils.hash_string(url)

    @property
    def url(self):
        return self._url

    def _with_feed_meta(self):
        return utils.ContextJSON(self.feed_repo / "__meta__.json")

    @property
    def feed_info(self) -> dict:
        return self._with_feed_meta().data

    @property
    def posts(self):
        ret = []
        for post in self.feed_repo.iterdir():
            if post.name.startswith("__"):
                continue
            ret.append(utils.ContextJSON(post))
        ret.sort(key=lambda x: x.name)
        ret.reverse()
        return ret

    def update_meta(self, title="", subtitle="", icon=None, **kwargs):
        with self._with_feed_meta() as d:
            for k, v in kwargs.items():
                d[k] = v
            d['url'] = self._url
            d['title'] = title
            d['subtitle'] = subtitle
            d['icon'] = icon
            d['last_updated'] = int(time.time())

    def update_post(
        self,
        article_url: str,
        published_unix: int,
        article_content: str = "",
        audio_url: Optional[str] = None,
        author: Optional[str] = None,
        icon=None,
        summary="",
        title="",
        video_url: Optional[str] = None,
        **kwargs
    ):
        with utils.ContextJSON(self.feed_repo / f"{published_unix}.json") as d:
            for (k, v) in kwargs.items():
                d[k] = v
            d['article_url'] = article_url
            d['audio_url'] = audio_url
            d['video_url'] = video_url
            d['author'] = video_url
            d['article_content'] = article_content
            d['published'] = published_unix
            d['title'] = title
            d['summary'] = summary
            d['icon'] = icon
