from wallabag.api.add_entry import AddEntry, Params as AddEntryParams
from wallabag.entry import Entry
from wallabag.config import Configs
from pathlib import Path
import json
from concurrent.futures import ThreadPoolExecutor
from tqdm import tqdm
import random
import sys
import contextlib
import io

pocket_fetched = Path('.').parent / "pocket_fetched"
config_file = Path.home() / ".config/wallabag-cli/config.ini"

config = Configs(config_file)

def find_urls():
    for bundle in pocket_fetched.glob('*.json'):
        data = json.loads(bundle.read_text())
        if data.get('error') is not None:
            bundle.unlink()
            continue
        for post in data['list'].values():
            print(post)
            url = post.get('resolved_url', post.get('given_url'))
            if url is None:
                continue
            yield AddEntry(config, url, {
                AddEntryParams.READ: post['status'] == 1,
                AddEntryParams.TITLE: post['resolved_title'],
                AddEntryParams.STARRED: post['favorite'] == 1
            })
    
data = find_urls()
data  = list(data)
# print(data[:4])
# exit(0)
random.shuffle(data)
with tqdm(total=len(data), desc="Ingerindo artigos", miniters=1, file=sys.stdout) as ops:
    def ingest_once(item):
        try:
            ops.update()
            ops.refresh(nolock=True)
            sys.stdout.flush()
            entry = item.request().response
            entry = Entry(entry)
            return entry
        except Exception as e:
            # pass
            print(e, file=sys.stdout)

    with ThreadPoolExecutor(max_workers=16) as tp:
        for item in tp.map(ingest_once, data):
            pass
            # if item is not None:
            #     ops.set_description(f"Ingerido {item.url}")
    print(folders)

