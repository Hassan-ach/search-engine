use crate::core::indexer::Page;
use std::collections::HashMap;
use std::env;
use std::error::Error;
use std::sync::OnceLock;

use sqlx::{query, Pool, Postgres};
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
    let mut tx = EXECUTER.get().unwrap().begin().await?;

    // Create a query type mapping
    let query = sqlx::query_as::<_, Page>(
        "WITH cte AS (
             SELECT id, url, html
             FROM pages
             WHERE indexed = FALSE
             FOR UPDATE SKIP LOCKED
             LIMIT 1
        )

        UPDATE pages
        SET indexed = TRUE
        FROM cte
        WHERE pages.id = cte.id
        RETURNING pages.id, pages.url, pages.html",
    );

    // Fetch Optional row
    let page = query.fetch_one(EXECUTER.get().unwrap()).await?;

    tx.commit().await?;

    Ok(page)
}

pub async fn batch_words(words: HashMap<String, u32>, page_id: Uuid) {
    for (word, count) in words.into_iter() {
        let word_id = match get_id_or_insert(&word).await {
            Some(id) => id,
            None => {
                warn!(word = %word, count = count,"no id found");
                continue;
            }
        };

        let res = query!(
            "INSERT INTO page_word(page_id, word_id, tf) VALUES ($1, $2, $3)",
            page_id,
            word_id,
            count as i32
        )
        .execute(EXECUTER.get().unwrap())
        .await;

        if let Err(err) = res {
            warn!(?err, word = %word, count = count, "Failed to index word");
        }
    }
}

async fn get_id_or_insert(word: &str) -> Option<Uuid> {
    let row = query!("SELECT id FROM words WHERE word = $1", word)
        .fetch_optional(EXECUTER.get().unwrap())
        .await
        .ok()?;

    let id = if let Some(row) = row {
        row.id
    } else {
        let inserted = query!("INSERT INTO words (word) VALUES ($1) RETURNING id", word)
            .fetch_one(EXECUTER.get().unwrap())
            .await;
        match inserted {
            Ok(inset) => inset.id,
            Err(err) => {
                warn!(?err,word= %word, "faild to insert word");
                return None;
            }
        }
    };

    Some(id)
}
