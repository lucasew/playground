from typing import List, Iterator
from dataclasses import field, dataclass
from .utils import fingerprint_string


@dataclass(kw_only=True)
class PodcastEpisode:
    title: str
    summary: str

    duration: int

    published: int # unix timestamp

    @property
    def payload(self):
        raise Exception("base class, not implemeneted")

    @property
    def thumbnail(self):
        raise Exception("base class, not implemented")

    @property
    def fingerprint(self):
        return [fingerprint_string(self.title), fingerprint_string(self.summary)]


@dataclass(kw_only=True)
class PodcastSource:
    title: str
    summary: str

    feed_url: str

    last_updated: int # unix timestamp

    @property
    def episodes(self) -> Iterator[PodcastEpisode]:
        pass

    @property
    def thumbnail(self):
        raise Exception("base class, not implemented")


    @property
    def fingerprint(self):
        return [fingerprint_string(self.title), fingerprint_string(self.summary)]


