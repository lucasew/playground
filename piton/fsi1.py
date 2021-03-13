#! /usr/bin/env nix-shell
#! nix-shell -i python3 -p python3

import collections
import timeit

class Graph:
    graph = {}
    def add_edge(self, a, b, cost):
        if not a in self.graph:
            self.graph[a] = []
        if not b in self.graph:
            self.graph[b] = []
        self.graph[a].append((b, cost))
        self.graph[b].append((a, cost))
    def edges_from(self, place):
        return self.graph[place]
    def assert_key(self, key):
        if key not in self.places():
            raise "undefined place specified"
    def places(self):
        return self.graph.keys()
    def bfs(self, root):
        self.assert_key(root)
        visited, queue = set(), collections.deque([root])
        visited.add(root)
        while queue:
            cur = queue.popleft()
            for neigh in self.graph[cur]:
                if neigh[0] not in visited:
                    visited.add(neigh[0])
                    queue.append(neigh[0])
        return visited
    def dfs(self, root):
        self.assert_key(root)
        for edge in self.edges_from(root):
            self.dfs(edge[0])
    def dfsl(self, root, target, limit):
        if limit == 0:
            return None
        self.assert_key(root)
        for edge in self.edges_from(root):
            if root == target:
                return [root]
            ret = self.dfsl(edge[0], target, limit - 1)
            if ret == None:
                continue
            else:
                return [root, *ret]
    def dis(self, root, target):
        i = 1
        while True:
            res = self.dfsl(root, target, i)
            if res != None:
                return res
            i+=1

def seedGraph():
    g = Graph()
    g.add_edge("club", "home", 1200)
    g.add_edge("club", "school", 1400)
    g.add_edge("club", "bank", 1100)
    g.add_edge("home", "bank", 1300)
    g.add_edge("home", "museum", 2700)
    g.add_edge("bank", "museum", 1700)
    g.add_edge("school", "bank", 1000)
    g.add_edge("school", "lake", 900)
    g.add_edge("bank", "lake", 1000)
    g.add_edge("bank", "park", 1800)
    g.add_edge("museum", "park", 1600)
    g.add_edge("lake", "park", 1600)
    return g

def romeniaGraph():
    g = Graph()
    g.add_edge("arad", "zerind", 75)
    g.add_edge("zerind", "oradea", 71)
    g.add_edge("arad", "timissora", 118)
    g.add_edge("arad", "sibiu", 140)
    g.add_edge("timissora", "lugoj", 111)
    g.add_edge("mehadia", "lugoj", 70)
    g.add_edge("oradea", "sibiu", 151)
    g.add_edge("mehadia", "dobreta", 75)
    g.add_edge("craiova", "dobreta", 75)
    g.add_edge("rimnicu vilcea", "sibiu", 80)
    g.add_edge("fagaras", "sibiu", 99)
    g.add_edge("rimnicu vilcea", "craiova", 146)
    g.add_edge("rimnicu vilcea", "pitesti", 97)
    g.add_edge("pitesti", "craiova", 138)
    g.add_edge("pitesti", "bucharest", 101)
    g.add_edge("giurgiu", "bucharest", 90)
    g.add_edge("urziceni", "bucharest", 85)
    g.add_edge("urziceni", "hirsova", 98)
    g.add_edge("eforie", "hirsova", 86)
    g.add_edge("urziceni", "vaslui", 142)
    g.add_edge("lasi", "vaslui", 92)
    g.add_edge("lasi", "neamt", 87)
    return g

def call(fn, *args):
    return lambda: fn(*args)
def benchmark(fn, *args):
    return timeit.timeit(call(fn, *args), number=100)

if __name__ == "__main__":
    g = romeniaGraph()
    print(g.graph.keys())
    print(g.dfsl("arad", "bucharest", 2))
    print(g.dfsl("arad", "bucharest", 4))
    print(g.dfsl("arad", "bucharest", 7))
    print(g.dis("arad", "bucharest"))
