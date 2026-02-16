use crate::core::indexer::Page;
use std::collections::HashMap;
use std::env;
use std::error::Error;
use std::sync::OnceLock;

use sqlx::{Pool, Postgres};
use tracing::{info, warn};
use uuid::Uuid;

// Function connect to postgres and test it
// return a pool connect
async fn db_connectioon() -> Result<Pool<Postgres>, Box<dyn Error>> {
    let host = env::var("PG_HOST").expect("PG_HOST must be set");
    let port = env::var("PG_PORT").expect("PG_PORT must be set");
    let user = env::var("PG_USER").expect("PG_USER must be set");
    let password = env::var("PG_PASSWORD").expect("PG_PASSWORD must be set");
    let dbname = env::var("PG_DBNAME").expect("PG_DBNAME must be set");

    let url = format!("postgres://{user}:{password}@{host}:{port}/{dbname}");
    let pool = sqlx::postgres::PgPool::connect(&url).await?;

    let _ = sqlx::query("SELECT 1 + 1 as sum").fetch_one(&pool).await?;

    info!("Data base connected successfully");
    Ok(pool)
}

pub static EXECUTER: OnceLock<Pool<Postgres>> = OnceLock::new();

// Function to initialisation of Executer global variable
// This function should be call once at programme life time
pub async fn init() {
    if let Ok(cnc) = db_connectioon().await {
        EXECUTER.set(cnc).unwrap();
    }
}

pub async fn get_page() -> Result<Page, Box<dyn Error>> {
    let tx = EXECUTER.get().unwrap().begin().await?;

    // Create a query type mapping
    let query = sqlx::query_as::<_, Page>(
        "WITH cte AS (
             SELECT id, url_id, html
             FROM pages
             WHERE indexed = FALSE
             FOR UPDATE SKIP LOCKED
             LIMIT 1
        )

        UPDATE pages
        SET indexed = TRUE
        FROM cte
        WHERE pages.id = cte.id
        RETURNING pages.id, pages.url_id, pages.html",
    );

    // Fetch Optional row
    let page = query.fetch_one(EXECUTER.get().unwrap()).await?;

    tx.commit().await?;

    Ok(page)
}

pub async fn batch_words(words: HashMap<String, u32>, page_id: Uuid) {
    if words.is_empty() {
        warn!(page_id = %page_id, "no words to index");
        return;
    }

    let map = match upsert_words(words.clone().into_keys().collect()).await {
        Some(m) => m,
        None => {
            warn!(page_id = %page_id, "failed to upsert words");
            return;
        }
    };

    let word_id_count: HashMap<Uuid, u32> = map
        .into_iter()
        .filter_map(|(word, id)| words.get(&word).map(|count| (id, *count)))
        .collect();

    if let Err(err) = link_words_to_page(page_id, word_id_count).await {
        warn!(?err, page_id = %page_id, "failed to link words to page");
    }
}

async fn link_words_to_page(
    page_id: Uuid,
    word_id_count: HashMap<Uuid, u32>,
) -> Result<(), sqlx::Error> {
    if word_id_count.is_empty() {
        return Ok(());
    }

    let mut word_ids = Vec::with_capacity(word_id_count.len());
    let mut counts = Vec::with_capacity(word_id_count.len());

    for (id, count) in word_id_count {
        word_ids.push(id);
        counts.push(count as i32);
    }

    sqlx::query!(
        r#"
        INSERT INTO page_word (page_id, word_id, tf)
        SELECT $1, * FROM UNNEST($2::uuid[], $3::int4[])
        ON CONFLICT (page_id, word_id) DO NOTHING
        "#,
        page_id,
        &word_ids,
        &counts
    )
    .execute(EXECUTER.get().expect("DB not initialized"))
    .await?;

    Ok(())
}

// Function to insert words in batch and return their ids
async fn upsert_words(words: Vec<String>) -> Option<HashMap<String, Uuid>> {
    if words.is_empty() {
        return Some(HashMap::new());
    }

    // Using UNNEST to pass the entire vector as one parameter ($1)
    let rows = sqlx::query!(
        r#"
        INSERT INTO words (word)
        SELECT * FROM UNNEST($1::text[])
        ON CONFLICT (word) DO UPDATE 
            SET word = EXCLUDED.word
        RETURNING id, word
        "#,
        &words[..]
    )
    .fetch_all(EXECUTER.get().expect("DB not initialized"))
    .await
    .ok()?;

    let mut ids = HashMap::with_capacity(rows.len());
    for row in rows {
        ids.insert(row.word, row.id);
    }

    Some(ids)
}
