#!/usr/bin/env nix-shell
#! nix-shell -p python3 -i python

from json import loads

JSON_FILE="/tmp/rem.json"
EDIT_LATER="pETqJncd8h6wqDu8N"

raw_data = None
with open(JSON_FILE) as f:
    raw_data = loads(f.read())

def normalize_node(node):
    if node.get("deletedAt") != None:
        return
    text = node["key"][0] if len(node["key"]) >= 1 else ""
    created_at = node["createdAt"]
    url = None
    try:
        url = node["crt"]["o"]["o"]["v"][0]["url"]
    except KeyError:
        pass
    except TypeError:
        pass
    return {
        "text": text,
        "created_at": created_at,
        "url": url
    }

indexed_items = {}
for item in raw_data["docs"]:
    indexed_items[item["_id"]] = item

edit_later_node = indexed_items[EDIT_LATER]
edit_later_children_ids = list(edit_later_node["typeChildren"])

edit_later_nodes = {}
for child in edit_later_children_ids:
    node = normalize_node(indexed_items[child])
    if node == None:
        continue
    edit_later_nodes[child] = node

first_edit_later_item = indexed_items[edit_later_children_ids[0]]
first_edit_later_item_normalized = normalize_node(first_edit_later_item)

list_edit_later = sorted(list(edit_later_nodes.values()), key = lambda v: v["created_at"])

labels = []
for item in list_edit_later:
    print(f"{item['text']} - {item['url']}")
