from typing import Optional
from time import time, mktime
from ..model import PodcastSource, PodcastEpisode
from ..utils import strduration2seconds, reexport_url
from dataclasses import dataclass

@dataclass
class RSSPodcastSource(PodcastSource):

    @property
    def thumbnail(self):
        return reexport_url(self._thumbnail_url)

    @property
    def refetch(self):
        return RSSPodcastSource.from_url(self.feed_url)

    @property
    def episodes(self):
        return iter(self._episodes)

    @staticmethod
    def from_feedparser_output(feedparser_payload):
        feed = feedparser_payload['feed']
        feed_url = None
        for link in feed['links']:
            if link['type'] in [ 'application/rss+xml' ]:
                feed_url = link['href']
        ret = RSSPodcastSource(
            title=feed['title'],
            summary=feed['content'],
            feed_url=feed_url,
            last_updated=int(time()),
        )
        ret._thumbnail_url = feed['image']['href'], # if feed.get('image') is not None else None,
        ret._episodes = []
        for episode in feedparser_payload['entries']:
            ret._episodes.append(RSSPodcastEpisode.from_feedparser_output(episode))
        return ret


    @staticmethod
    def from_url(*feedparser_args, **feedparser_kwargs):
        from feedparser import parse as feedparse
        feedparser_payload = feedparse(*feedparser_args, **feedparser_kwargs)
        return RSSPodcastSource.from_feedparser_output(feedparser_payload)

@dataclass
class RSSPodcastEpisode(PodcastEpisode):

    @property
    def payload(self):
        return reexport_url(self._payload_url)

    @property
    def thumbnail(self):
        return reexport_url(self._thumbnail_url)

    @staticmethod
    def from_feedparser_output(feedparser_payload):
        duration = 0
        if feedparser_payload.get('itunes_duration') is not None:
            duration = strduration2seconds(feedparser_payload['itunes_duration'])
        ret = RSSPodcastEpisode(
            title=feedparser_payload['title'],
            summary=feedparser_payload['summary'],
            duration=duration,
            published=int(mktime(feedparser_payload['published_parsed']))
        )
        ret._thumbnail_url = feedparser_payload['image']['href'] # if episode.get('image') is not None else None,
        for link in feedparser_payload['links']:
            if link['type'].startswith('audio'):
                ret._payload_url = link['href']
        return ret

