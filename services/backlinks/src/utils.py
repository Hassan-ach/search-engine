# In this module, we define a function to calculate the Inverse Document Frequency (IDF)
# and update it in the 'words' table of a PostgreSQL database in a single batch operation.
# and a function to establish a connection to the PostgreSQL database using environment variables.

import os
import psycopg2
import uuid
import numpy as np
from psycopg2.extras import execute_values


class NodeMapper:
    def __init__(self):
       self. node_id = {}
       self. node_uuid = {}
       self. next_id = 0

    def get_id(self, u: uuid.UUID) -> int:
        if u not in self.node_id:
            self.node_id[u] = self.next_id
            self.node_uuid[self.next_id] = u
            self.next_id += 1
        return self.node_id[u]

    def  get_uuid(self, id: int) -> uuid.UUID:
        return self.node_uuid[id]




def get_connection():
    return psycopg2.connect(
        host=os.getenv('PG_HOST'),
        port=os.getenv('PG_PORT'),
        database=os.getenv('PG_DBNAME'),
        user=os.getenv('PG_USER'),
        password=os.getenv('PG_PASSWORD')
    )

def get_graph_edges(m : NodeMapper) -> list[tuple[int, int]]:
    print("Fetching graph edges from database...")
    conn = get_connection()
    try:
        with conn.cursor() as cursor:
            cursor.execute(
                """
                SELECT from_url, to_url
                FROM graph_edges;
                """
            )
            print("Graph edges fetched successfully. N =", cursor.rowcount)
            return [(m.get_id(f),  m.get_id(t)) for f, t in cursor.fetchall()]
    except Exception as _:
        raise
    finally:
        conn.close()



def persist_pagerank(m: NodeMapper, pagerank: np.ndarray):
    print("Persisting PageRank scores to database...")
    conn = get_connection()
    try:
        with conn.cursor() as cursor:
            data = [
                (m.get_uuid(i), float(pr))
                for i, pr in enumerate(pagerank)
            ]

            execute_values(
                cursor,
                """
                INSERT INTO page_rank (url_id, score)
                VALUES %s
                ON CONFLICT (url_id) DO UPDATE
                SET score = EXCLUDED.score
                """,
                data,
                page_size=10_000  
            )

        conn.commit()
        print("PageRank scores persisted successfully.")
    finally:
        conn.close()

