from wallabag.api.add_entry import AddEntry, Params as AddEntryParams
from wallabag.entry import Entry
from wallabag.config import Configs
from pathlib import Path
import csv
from concurrent.futures import ThreadPoolExecutor
from tqdm import tqdm
import random
import sys
import contextlib
import io

config_file = Path.home() / ".config/wallabag-cli/config.ini"

config = Configs(config_file)

url = "https://google.com"
starred = False
read = False
# print(entry)

folders = set()

with open("instapaper-export.csv", 'r') as f:
    data = csv.DictReader(f)
    data  = list(data)
    random.shuffle(data)
    with tqdm(total=len(data), desc="Ingerindo artigos", miniters=1, file=sys.stdout) as ops:
        def ingest_once(item):
            try:
                # print('item', item)
                folder = item['Folder']
                title = item['Title']
                url = item['URL']
                ops.update()
                ops.refresh(nolock=True)
                sys.stdout.flush()
                entry = Entry(AddEntry(config, url, {
                    AddEntryParams.READ: folder == 'Archive',
                    AddEntryParams.TITLE: title
                #     AddEntryParams.STARRED: starred
                }).request().response)
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

