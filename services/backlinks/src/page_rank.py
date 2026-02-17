# in this module we implement the PageRank algorithm
# to compute the importance of web pages based on their backlinks.

from collections import defaultdict
import numpy as np


def build_adjacency_list(edges: list[tuple[int, int]]) -> tuple[dict[int, list[int]], dict[int, int]]:
    adj = defaultdict(list)
    out_degree = defaultdict(int)
    for u, v in edges:
        out_degree[u] += 1
        adj[u].append(v)
    return (adj, out_degree)

def build_stochastic_matrix(adj: dict[int, list[int]], out_degree: dict[int, int]) -> np.ndarray:
    max_node = max(
        max(adj.keys(), default=0),
        max((v for vs in adj.values() for v in vs), default=0)
    )
    N = max_node + 1

    M = np.zeros((N, N))
    for u in adj:
        for v in adj[u]:
            M[v][u] = 1 / out_degree[u]
    return M


def pagerank(edges: list[tuple[int, int]], d: float = 0.85, a=1e-5) -> np.ndarray:
    print("Calculating PageRank...")

    adj, out_degree = build_adjacency_list(edges)
    M = build_stochastic_matrix(adj, out_degree)

    N = M.shape[1]
    w = np.ones(N) / N
    M_hat = d * M
    v = M_hat @ w + (1 - d) / N
    while np.linalg.norm(w - v) >= a:
        w = v
        v = M_hat @ w + (1 - d) / N

    print("PageRank calculation completed.")
    return v

