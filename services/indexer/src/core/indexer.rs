use std::thread::sleep;
use std::time::Duration;

use crate::core::psql::*;
use crate::core::text_sink::parse;
use sqlx::prelude::FromRow;
use tracing::warn;
use uuid::Uuid;

#[derive(Debug, FromRow)]
pub struct Page {
    id: Uuid,
    url_id: Uuid,
    html: String,
}

pub async fn index() {
    loop {
        match get_page().await {
            Ok(page) => {
                match parse(page.html).await {
                    Ok(words) => {
                        batch_words(words, page.id).await;
                    }
                    Err(err) => {
                        warn!(?err, url = %page.url_id, "faild to parse a page");
                    }
                };
            }
            Err(err) => {
                warn!(?err, "failed to get page");
                sleep(Duration::from_secs(5));
            }
        };
    }
}
