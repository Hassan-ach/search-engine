# in this module we implement the PageRank algorithm
# to compute the importance of web pages based on their backlinks.

from collections import defaultdict
import numpy as np
import logging
from scipy.sparse import csr_matrix, lil_matrix

logger = logging.getLogger(__name__)


def build_adjacency_list(edges: list[tuple[int, int]]) -> tuple[dict[int, list[int]], dict[int, int]]:
    adj = defaultdict(list)
    out_degree = defaultdict(int)
    for u, v in edges:
        out_degree[u] += 1
        adj[u].append(v)
    return adj, out_degree


def build_stochastic_matrix(
    adj: dict[int, list[int]],
    out_degree: dict[int, int],
    N: int
) -> csr_matrix:
    M = lil_matrix((N, N))
    for u in adj:
        for v in adj[u]:
            M[v, u] = 1 / out_degree[u]
    return M.tocsr()


def pagerank(
    edges: list[tuple[int, int]],
    N: int | None = None,
    d: float = 0.85,
    tol: float = 1e-5,
    max_iterations: int = 100,
    personalization: np.ndarray | None = None,
) -> np.ndarray:
    logger.info("Calculating PageRank...")

    adj, out_degree = build_adjacency_list(edges)

    if N is None:
        max_node = max(
            max(adj.keys(), default=0),
            max((v for vs in adj.values() for v in vs), default=0)
        )
        N = max_node + 1

    M = build_stochastic_matrix(adj, out_degree, N)

    # Dangling nodes: no outgoing edges, not column-stochastic without this
    dangling_nodes = np.array([n for n in range(N) if n not in out_degree], dtype=int)

    # Teleportation vector — uniform by default, or personalized
    if personalization is None:
        p = np.ones(N) / N
    else:
        if personalization.shape[0] != N:
            raise ValueError(f"personalization vector length {personalization.shape[0]} != N={N}")
        p = personalization / personalization.sum()

    w = np.ones(N) / N

    for iterations in range(max_iterations):
        dangling_contrib = d * w[dangling_nodes].sum() * p
        v = d * (M @ w) + dangling_contrib + (1 - d) * p

        residual = np.linalg.norm(w - v, ord=1)
        w = v

        if residual < tol:
            logger.info(f"Converged in {iterations + 1} iterations.")
            break
    else:
        logger.warning(
            f"Did not converge after {max_iterations} iterations. "
            f"Residual: {residual:.2e}"
        )

    return v
