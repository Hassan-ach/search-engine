mod core;
use dotenv::from_path;
use std::{error::Error, fs};
use tracing::{level_filters::LevelFilter, Level};

use tracing_subscriber::{filter::Targets, fmt, layer::SubscriberExt, util::SubscriberInitExt};

use crate::core::indexer::index;

async fn init() -> Result<(), Box<dyn Error>> {
    //
    from_path("../../.env").expect("cannot open .env file");
    // Print the environment variables
    println!(
        "PG_HOST={}",
        std::env::var("PG_HOST").unwrap_or("not set".to_string())
    );
    println!(
        "PG_PORT={}",
        std::env::var("PG_PORT").unwrap_or("not set".to_string())
    );
    println!(
        "PG_USER={}",
        std::env::var("PG_USER").unwrap_or("not set".to_string())
    );
    println!(
        "PG_PASSWORD={}",
        std::env::var("PG_PASSWORD").unwrap_or("not set".to_string())
    );
    println!(
        "PG_DBNAME={}",
        std::env::var("PG_DBNAME").unwrap_or("not set".to_string())
    );

    let log_file = fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open("indexer.log")
        .expect("cannot open log file");

    let writer = std::sync::Mutex::new(log_file);

    let filter = Targets::new()
        .with_target("my_crate", Level::INFO)
        .with_target("other_crate", LevelFilter::OFF);

    tracing_subscriber::registry()
        .with(fmt::layer().with_writer(writer).with_ansi(false).json())
        .with(fmt::layer().with_writer(std::io::stdout).with_ansi(true))
        .with(filter)
        .init();
    crate::core::psql::init().await;
    Ok(())
}

// TODO:
// [] fix logger.
// [] impl concurrency.
#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    init().await?;
    println!("Staring Indexing...");
    index().await;
    Ok(())
}
