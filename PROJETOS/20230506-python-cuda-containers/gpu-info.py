#!/usr/bin/env python3

from subprocess import run, PIPE
from sys import stderr
from pathlib import Path

def run_lines(command):
    return run(command, stdout=PIPE, stderr=stderr).stdout.decode('utf-8').split('\n')

def split_multiple(text, splitter=" "):
    return [item for item in text.split(splitter) if item != '']

def is_pid_exists(pid: int):
    p = Path("/proc") / str(pid)
    return p.exists()

cuda_pids = set([int(x) for x in run_lines(["nvidia-smi", "--query-compute-apps=pid", "--format=csv,noheader"]) if len(x) > 0 and is_pid_exists(x) ])

container_pids = set()
for container in run_lines(["docker", "ps", "--format", "{{.ID}}"]):
    top_lines = [split_multiple(line)[1] for line in run_lines(["docker", "top", container])[1:] if len(line) > 0]
    for line in top_lines:
        if not is_pid_exists(line):
            continue
        container_pids.add(int(line))
    #print('container', container, top_lines)

cuda_out_of_container = cuda_pids - container_pids
cuda_intersection = cuda_pids.intersection(container_pids)
is_all_cuda_in_containers = cuda_pids.issubset(container_pids) 
print(dict(
    cuda_out_of_container=cuda_out_of_container,
    cuda_intersection=cuda_intersection,
    is_all_cuda_in_containers=is_all_cuda_in_containers
))
