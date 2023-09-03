from mediakit_project.utils.feed import get_root_dir
from mediakit_project import utils
import logging

COMMAND_DESCRIPTION = "Extract links for extractors from JSON posts"

logger = logging.getLogger(__name__)


def command(parser):
    def handle(args):
        for feed_dir in get_root_dir().iterdir():
            if feed_dir.name.startswith("__"):
                continue
            for post in feed_dir.iterdir():
                if post.name.startswith("__"):
                    continue
                logger.debug(_("Processing {item}"). format(item=str(post)))
                post_data = utils.ContextJSON(post / "__meta__.json").data
                article_url = post_data['article_url']
                video_url = post_data.get('video_url')
                audio_url = post_data.get('audio_url')
                if video_url is None:
                    video_url = article_url
                if audio_url is None:
                    audio_url = video_url  # audio can be extracted from video
                (post / "__article_url").write_text(article_url)
                (post / "__audio_url").write_text(audio_url)
                (post / "__video_url").write_text(video_url)
    return handle
