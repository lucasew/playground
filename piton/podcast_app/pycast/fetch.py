from json import dumps
from dataclasses import dataclass, field
from typing import List

def fingerprint_string(s: str) -> str:
    from hashlib import sha256
    symboldict = [
        *map(chr, range(ord('A'), ord('Z') + 1)),
        *map(chr, range(ord('0'), ord('9') + 1)),
    ]
    m = sha256()
    l = []
    for char in s.upper():
        if char in symboldict:
            l.append(char)
    txt = "".join(l)
    m.update(txt.encode())
    return m.hexdigest()

def strduration2seconds(duration: str) -> int:
    parts = duration.split(":")
    multiplier = 1
    ret = 0
    while len(parts) > 0:
        part = int(parts.pop())
        ret += multiplier * part
        multiplier *= 60
    return ret


@dataclass
class PodcastEpisode:
    title: str
    summary: str

    thumbnail_url: str
    payload_url: str

    duration: int

    published: int # unix timestamp

    @property
    def fingerprint(self):
        return [fingerprint_string(self.title), fingerprint_string(self.summary)]


@dataclass
class PodcastSource:
    title: str
    summary: str

    thumbnail_url: str
    feed_url: str

    last_updated: int # unix timestamp

    episodes: List[PodcastEpisode] = field(default_factory=list)

    @property
    def fingerprint(self):
        return [fingerprint_string(self.title), fingerprint_string(self.summary)]

    @staticmethod
    def from_feed_url(url: str):
        from feedparser import parse
        from time import time, mktime
        parsed = parse(url)
        feed = parsed['feed']
        episodes = []
        for episode in parsed['entries']:
            payload_url = None
            for link in episode['links']:
                if link['type'].startswith('audio'):
                    payload_url = link['href']
            duration = 0
            if episode.get('itunes_duration') is not None:
                duration = strduration2seconds(episode['itunes_duration'])
            if payload_url is None:
                continue
            episodes.append(PodcastEpisode(
                title=episode['title'],
                summary=episode['summary'],
                thumbnail_url=episode['image']['href'] if episode.get('image') is not None else None,
                payload_url=payload_url,
                duration=duration,
                published=mktime(episode['published_parsed'])
            ))

        ret = PodcastSource(
            title=feed['title'],
            summary=feed['content'],
            thumbnail_url=feed['image']['href'] if feed.get('image') is not None else None,
            feed_url=url,
            last_updated=int(time()),
            episodes=episodes
        )
        return ret


item = PodcastSource.from_feed_url("http://feeds.libsyn.com/73434/rss")
print(item)
print(item.fingerprint)

assert(strduration2seconds("2:03"), 123)
assert(strduration2seconds("1:2:03"), 123 + 3600)
