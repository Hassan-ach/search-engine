use tokio::time::sleep;
use tokio::time::Duration;

use crate::core::psql::*;
use crate::core::text_sink::parse;
use sqlx::prelude::FromRow;
use uuid::Uuid;

#[derive(Debug, FromRow)]
pub struct Page {
    id: Uuid,
    url_id: Uuid,
    html: String,
}

pub async fn index() {
    match get_page().await {
        Ok(page) => {
            match parse(page.html).await {
                Ok(words) => {
                    batch_words(words, page.id).await;
                }
                Err(err) => {
                    println!("parse page id:{},  err: {}", page.url_id, err);
                }
            };
        }
        Err(err) => {
            println!("get page err: {}", err);
            sleep(Duration::from_millis(50)).await;
        }
    };
}
