import logging
import time
from psql import get_connection, release_connection, retry_on_db_error

logger = logging.getLogger(__name__)

@retry_on_db_error(max_retries=3, delay=1.0, backoff=2.0)
def idf():
    start_time = time.time()
    logger.info("Running IDF calculation...")
    conn = get_connection()
    try:
        with conn.cursor() as cursor:
            cursor.execute("""
                UPDATE words
                SET idf = LOG((SELECT COUNT(*) FROM pages) / (1 + sub.df))
                FROM (
                    SELECT word_id, COUNT(DISTINCT page_id) AS df 
                    FROM page_word 
                    GROUP BY word_id
                ) sub
                WHERE words.id = sub.word_id
            """)
            affected_rows = cursor.rowcount
            conn.commit()
            duration = time.time() - start_time
            logger.info(f"IDF updated for {affected_rows} words in {duration:.2f}s")
    except Exception as e:
        logger.error(f"IDF calculation failed: {e}", exc_info=True)
        conn.rollback()
        raise
    finally:
        release_connection(conn)
