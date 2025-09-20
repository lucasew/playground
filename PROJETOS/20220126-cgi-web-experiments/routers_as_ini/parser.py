#!/usr/bin/env python3

from configparser import ConfigParser
cfg = ConfigParser()
cfg.read("test.ini")

print(cfg.sections())
