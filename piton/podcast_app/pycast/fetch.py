from json import dumps
from dataclasses import dataclass, field
from typing import List
from .utils import fingerprint_string, strduration2seconds
from .model import PodcastEpisode, PodcastSource
from .extractor import RSSPodcastSource

item = RSSPodcastSource.from_url("http://feeds.libsyn.com/73434/rss")
print(item)
print(item.fingerprint)

assert strduration2seconds("2:03"), 123
assert strduration2seconds("1:2:03"), 123 + 3600
