use crate::core::psql::*;
use crate::core::text_sink::parse;
use sqlx::prelude::FromRow;
use tracing::warn;
use uuid::Uuid;

#[derive(Debug, FromRow)]
pub struct Page {
    id: Uuid,
    url: String,
    html: String,
}

pub async fn index() {
    let page = match get_page().await {
        Ok(p) => p,
        Err(err) => {
            warn!(?err, "failed to get page");
            return;
        }
    };

    let words = match parse(page.html).await {
        Ok(w) => w,
        Err(err) => {
            warn!(?err, url = %page.url, "faild to parse a page");
            return;
        }
    };

    batch_words(words, page.id).await;
}
