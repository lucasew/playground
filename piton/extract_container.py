#!/usr/bin/env python3

from argparse import ArgumentParser
import json
from pathlib import Path
import urllib
from urllib.request import build_opener, Request,urlopen, _opener as __opener
from urllib.error import HTTPError
from pprint import pprint

parser = ArgumentParser()

# no namespace == library, ex: nginx -> library/nginx
parser.add_argument('image') 
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


def fetch_image(image='library/hello-world', os='linux', architecture='amd64'):
    token = json.load(urlopen(f"https://auth.docker.io/token?service=registry.docker.io&scope=repository:{image}:pull"))['token']
    
    urlopen("https://registry.hub.docker.com/v2")
    req = f"https://registry.hub.docker.com/v2/repositories/{image}/tags"
    with urlopen(req) as res:
        # print(res.headers['WWW-Authenticate'])
        data = json.load(res)
        for result in data['results']:
            for image_data in result['images']:
                if image_data['os'] == os and image_data['architecture'] == architecture:
                    digest = image_data['digest']
                    req = Request(f"https://registry.hub.docker.com/v2/{image}/manifests/{digest}", headers={
                        "Authorization": f"Bearer {token}",
                        "Accept": "application/vnd.oci.image.manifest.v1+json"
                    })
                    with urlopen(req) as res:
                        # res_str = res.read().decode('utf-8')
                        manifest = json.load(res)
                        
                        req = Request(f"https://registry.hub.docker.com/v2/{image}/blobs/{manifest['config']['digest']}", headers={
                            "Authorization": f"Bearer {token}",
                            # "Accept": "application/vnd.oci.image.config.v1+json"
                        })
                        with urlopen(req) as res:
                            print(res.status, res.headers)
                            res_str = res.read()
                            manifest['config'] = json.loads(res_str)

                        image_data['manifest'] = manifest
                        yield image_data
                            

image = next(fetch_image(image=args.image))

pprint(image)
