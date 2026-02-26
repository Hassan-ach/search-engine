mod core;

use crate::core::config::load_config;
use crate::core::indexer::Indexe;
use crate::core::indexer::Indexer;
use crate::core::psql::Psql;
use slog::{error, info, o, Drain, Logger};
use std::error::Error;
use std::fs::OpenOptions;
use std::io;
use std::path::Path;
use std::sync::Arc;
use tokio::task::JoinHandle;
use tokio::time::timeout;

pub fn init_logger(log_file: &str) -> io::Result<Logger> {
    // Create log directory
    if let Some(parent) = Path::new(log_file).parent() {
        std::fs::create_dir_all(parent)?;
    }

    // Terminal: Pretty, colored output
    let term = slog_term::TermDecorator::new()
        .stderr()
        .force_color()
        .build();
    let term_drain = slog_term::FullFormat::new(term)
        .use_local_timestamp()
        .build()
        .fuse();

    // File: JSON format
    let file = OpenOptions::new()
        .create(true)
        .append(true)
        .open(log_file)?;

    let json_drain = slog_json::Json::new(file)
        .add_default_keys()
        .add_key_value(slog::o!(
            "service" => "indexer",
        ))
        .build()
        .fuse();

    // Combine both - logs go to terminal AND file
    let drain = slog::Duplicate::new(term_drain, json_drain).fuse();

    // Wrap in Arc for thread-safe sharing across Tokio tasks
    // slog_async makes it non-blocking
    let async_drain = slog_async::Async::new(drain).chan_size(1024).build();

    Ok(Logger::root(Arc::new(async_drain).fuse(), o!()))
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let conf = load_config("../../.env".to_string());
    let log = init_logger(conf.app.log_path.clone().as_str())?;

    let psql = Psql::new(conf.psql, log.clone()).await?;
    let indx = Arc::new(Indexer::<Psql>::new(psql, conf.app, log.clone()));

    let token = tokio_util::sync::CancellationToken::new();

    let mut handles = Vec::<JoinHandle<()>>::new();

    for i in 0..4 {
        info!(log, "Starting indexer indexes"; "indexer_id" => i);
        let token_clone = token.clone();
        let job = indx.clone();
        handles.push(tokio::spawn(async move {
            job.start(token_clone).await;
        }));
    }

    let log_clone = log.clone();
    let sig_handle = tokio::spawn(async move {
        match tokio::signal::ctrl_c().await {
            Ok(()) => {
                info!(log_clone, "Ctrl-C received, sending shutdown signal");
                token.cancel();
            }
            Err(err) => {
                error!(log_clone, "Failed to listen for Ctrl-C"; "error" => %err);
            }
        }
    });

    match timeout(std::time::Duration::from_secs(30), async {
        for handle in handles {
            if let Err(err) = handle.await {
                error!(log, "Task failed to join"; "error" => %err);
            }
        }
    })
    .await
    {
        Ok(_) => info!(log, "All indexer tasks completed"),
        Err(_) => info!(
            log,
            "Timeout reached while waiting for indexer tasks to complete"
        ),
    }

    sig_handle.abort();
    info!(log, "Indexer service shutting down gracefully");

    Ok(())
}
