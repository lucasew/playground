#!/usr/bin/env python3

from argparse import ArgumentParser
import json
from pathlib import Path
import urllib
from urllib.request import build_opener, Request,urlopen, _opener as __opener
from urllib.error import HTTPError
from pprint import pprint
import tarfile
import tempfile

parser = ArgumentParser()

# no namespace == library, ex: nginx -> library/nginx
parser.add_argument('image') # example: library/hello-world
parser.add_argument('--os', default='linux')
parser.add_argument('--arch', default='amd64')
parser.add_argument('output')

args = parser.parse_args()

class QuietHTTPErrorProcessor(urllib.request.HTTPErrorProcessor):
    http_response = https_response = lambda self, request, response: response

opener = build_opener(QuietHTTPErrorProcessor)

def urlopen(*args,**kwargs):
    # there is the tradeoff
    # - raw urlopen can handle redirections (ex: the image config)
    # - that custom instance do not raise exceptions when non 2xx
    # yeah, that sucks but this is a workaround for that :)
    from urllib.request import urlopen as _open
    return _open(*args, **kwargs)
    # return opener.open(*args, **kwargs)


token = json.load(urlopen(f"https://auth.docker.io/token?service=registry.docker.io&scope=repository:{args.image}:pull"))['token']

def fetch_image_data():
    # urlopen("https://registry.hub.docker.com/v2")
    req = f"https://registry.hub.docker.com/v2/repositories/{args.image}/tags"
    with urlopen(req) as res:
        # print(res.headers['WWW-Authenticate'])
        data = json.load(res)
        for result in data['results']:
            for image_data in result['images']:
                if image_data['os'] == args.os and image_data['architecture'] == args.arch:
                    digest = image_data['digest']
                    req = Request(f"https://registry.hub.docker.com/v2/{args.image}/manifests/{digest}", headers={
                        "Authorization": f"Bearer {token}",
                        "Accept": "application/vnd.oci.image.manifest.v1+json"
                    })
                    with urlopen(req) as res:
                        manifest = json.load(res)
                        
                        req = Request(f"https://registry.hub.docker.com/v2/{args.image}/blobs/{manifest['config']['digest']}", headers={
                            "Authorization": f"Bearer {token}",
                        })
                        with urlopen(req) as res:
                            print(res.status, res.headers)
                            res_str = res.read()
                            manifest['config'] = json.loads(res_str)

                        image_data['manifest'] = manifest
                        yield image_data
                            
image = next(fetch_image_data())
pprint(image)

for layer in image['manifest']['layers']:
    print('layer', layer)

    req = Request(f"https://registry.hub.docker.com/v2/{args.image}/blobs/{layer['digest']}", headers={
        "Authorization": f"Bearer {token}",
    })
    with urlopen(req) as res:
        blob_size = int(res.getheader('Content-Length'))
        print(f"Extraindo blob {layer['digest']} ({blob_size}b)")
        with tempfile.TemporaryFile() as tempfile:
            while True:
                data = res.read(16*1024)
                if not data:
                    break
                tempfile.write(data)
            tempfile.seek(0)
            with tarfile.open(mode='r', fileobj=tempfile) as layer:
                layer.extractall(args.output)
    
