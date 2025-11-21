# In this module, we define a simple graph
# structure to represent backlinks between web pages.
# Each page is a node, and each backlink is a directed edge.

import random

class Graph:
    def __init__(self) -> None:
        self.graph = {}
        self.incoming = {}
        self.current = None
    
    def add_edge(self, u: str, v: str) -> None:
        if u not in self.graph:
            self.graph[u] = []

        if self.current is None:
            self.current = u
        self.graph[u].append(v)

        if v not in self.incoming:
            self.incoming[v] = []
        self.incoming[v].append(u)
    
    def get_randm_page(self) -> str|None:
        if self.current not in self.graph or not self.graph[self.current]:
            return None

        self.current = random.choice(self.graph[self.current])
        return self.current

    def get_links_count(self, u = None) -> int:
        if self.current not in self.graph or not self.graph[self.current]:
            return 0

        if u is None:
            u = self.current
            
        return len(self.graph.get(u, []))

    def get_incoming_links(self, u = None) -> int:
        if self.current not in self.incoming or not self.incoming[self.current]:
            return 0

        if u is None:
            u = self.current

        return len(self.incoming.get(u, []))
