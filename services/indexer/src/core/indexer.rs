use tokio::time::sleep;
use tokio::time::Duration;

use crate::core::config::AppConfig;
use crate::core::psql::DB;
use crate::core::text_sink::parse;
use sqlx::prelude::FromRow;
use tokio_util::sync::CancellationToken;
use uuid::Uuid;

#[derive(Debug, FromRow)]
pub struct Page {
    pub id: Uuid,
    pub url_id: Uuid,
    pub html: String,
}

#[derive(Debug, Clone)]
pub struct Indexer<DBImpl: DB> {
    db: DBImpl,
    conf: AppConfig,
}

pub trait Indexe {
    async fn start(&self, tk: CancellationToken);
}

impl<DBImpl: DB> Indexer<DBImpl> {
    pub fn new(db: DBImpl, conf: AppConfig) -> Self {
        Indexer { db, conf }
    }

    pub async fn index(&self, tk: &CancellationToken) {
        if tk.is_cancelled() {
            println!("indexing task received shutdown signal, stopping...");
            return;
        }
        match self.db.get_page().await {
            Ok(page) => {
                println!(
                    "get page id: {}, url_id: {}, html_size: {}",
                    page.id,
                    page.url_id,
                    page.html.len()
                );

                match parse(page.html).await {
                    Ok(words) => {
                        println!("parse page id:{}, word count: {}", page.url_id, words.len());
                        self.db.batch_words(words, page.id).await;
                    }
                    Err(err) => {
                        println!("parse page id:{},  err: {}", page.url_id, err);
                    }
                };
            }
            Err(err) => {
                println!("get page err: {}", err);
                tk.cancel();
                sleep(Duration::from_millis(300)).await;
            }
        };
    }
}

impl<DBImpl: DB> Indexe for Indexer<DBImpl> {
    async fn start(&self, tk: CancellationToken) {
        loop {
            tokio::select! {
                _ = tk.cancelled() => {
                    println!("shutdown signal received, stopping indexer...");
                    break;
                }
                _ = async {
                    self.index(&tk).await;
                    sleep(Duration::from_millis(self.conf.loop_delay_ms)).await;
                } => {}
            }
        }
    }
}
