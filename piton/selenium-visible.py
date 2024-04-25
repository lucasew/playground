#!/usr/bin/env -S sd nix shell
#!nix-shell -i python3 -p python3Packages.selenium python3Packages.opencv4 chromedriver

from selenium import webdriver
from selenium.webdriver.common.by import By
import subprocess
import time
import numpy as np
import cv2

service = webdriver.ChromeService(service_args=['--disable-build-check'], log_output=subprocess.STDOUT)

driver = webdriver.Chrome(service=service)

driver.get('https://google.com')
size = driver.get_window_size()

time.sleep(10)

depthmap = np.zeros((size['height'], size['width']), dtype='uint8')

bbxs = driver.execute_script('''
const vw = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)
const vh = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0)
let zIndexes = [
    {zIndex: 0, x: 0, y: 0, w: vw, h: vh}
]
function iterateOverNode(node, zThreshold) {
    if (!(node instanceof Element)) {
        return
    }
    console.log(node, zThreshold)
    const computedStyle = window.getComputedStyle(node)
    let zIndex = computedStyle.getPropertyValue('z-index')
    if (zIndex === 'auto') {
        zIndex = 0   
    }
    if (computedStyle.getPropertyValue('position') === 'absolute') {
        zIndex = 1
    }
    if (zIndex > zThreshold) {
        zIndexes.push({
            zIndex: zIndex,
            x: node.offsetLeft,
            y: node.offsetTop,
            w: node.offsetWidth,
            h: node.offsetHeight
        })
    }
    node.childNodes.forEach((node) => iterateOverNode(node, zIndex))
}
iterateOverNode(document.body, 0)
return zIndexes
''')

print('result', bbxs)
# time.sleep(60)
for bbx in bbxs:
    x = int(bbx['x'])
    y = int(bbx['y'])
    w = int(bbx['w'])
    h = int(bbx['h'])
    z_index = int(bbx['zIndex'])

    if x < 0:
        continue
    if y < 0:
        continue

    depthmap[y:y+h, x:x+w] = np.max([
        depthmap[y:y+h, x:x+w],
        np.array(np.ones((h, w)) * z_index * 32, dtype='uint8')
    ], axis=0)

print(np.max([
[1,2,3],
[3,2,1]
]
, axis=0))

cv2.imwrite("/tmp/teste_depthmap.png", depthmap)
driver.save_screenshot("/tmp/teste_screenshot.png")
