from utils import get_connection

def idf():
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
