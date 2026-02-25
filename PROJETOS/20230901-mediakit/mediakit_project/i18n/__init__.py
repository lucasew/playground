import gettext
import logging
from pathlib import Path

logger = logging.getLogger(__name__)

locale_dir = Path(__file__).parent

gettext.install(
    "mediakit_project",
    localedir=str(locale_dir)
)

gettext.gettext = _

logger.debug(_('Loading locale data from "{locale_dir}"').format(locale_dir=locale_dir))
