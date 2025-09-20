consumer_key = "preencher"
access_token = "preencher"

from pocket import Pocket
from pathlib import Path
import json

pocket = Pocket(consumer_key, access_token)

output_dir = Path('.').parent / "pocket_fetched"


def index_gen():
    offset = 0
    batch_size = 30
    while True:
        yield offset
        offset += batch_size
for offset in index_gen():
    file_name = output_dir / f"fetch_{offset:06}.json"
    if file_name.exists():
        continue
    print(f"fetch offset={offset}")
    res, _headers = pocket.get(
        images=1,
        videos=1,
        tags=1,
        rediscovery=1,
        annotations=1,
        authors=1,
        itemOptics=1,
        meta=1,
        posts=1,
        total=1,
        forceaccount=1,
        offset=offset,
        count=30,  # max count per request according to api docs
        state='all',
        sort='newest',
        detailType='complete',
    )
    if res.get('list') is not None and len(res['list']) == 0:
        break
    file_name.write_text(json.dumps(res))
