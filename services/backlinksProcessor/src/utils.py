# In this module, we define a function to calculate the Inverse Document Frequency (IDF)
# and update it in the 'words' table of a PostgreSQL database in a single batch operation.
# and a function to establish a connection to the PostgreSQL database using environment variables.

import os
import psycopg2
from dotenv import load_dotenv

load_dotenv("../../../.env")

def get_connection():
    return psycopg2.connect(
        host=os.getenv('PG_HOST'),
        port=os.getenv('PG_PORT'),
        database=os.getenv('PG_DATABASE'),
        user=os.getenv('PG_USER'),
        password=os.getenv('PG_PASSWORD')
    )

def calculate_idf_in_batch():
    conn = get_connection()
    try:
        with conn.cursor() as cursor:
            cursor.execute("""
                UPDATE words
                SET idf = (1 + LOG((SELECT COUNT(*) FROM pages)) * 1.0 / sub.df)
                FROM (
                    SELECT word_id, COUNT(DISTINCT page_id) AS df 
                    FROM page_word 
                    GROUP BY word_id
                ) sub
                WHERE words.id = sub.word_id
            """)
            conn.commit()
    except Exception as _:
        conn.rollback()
        raise
    finally:
        conn.close()
