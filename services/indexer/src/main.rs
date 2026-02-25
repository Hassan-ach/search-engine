mod core;
use crate::core::indexer::index;
use dotenv::from_path;
use std::error::Error;

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

    println!(
        "DATABASE_URL={}",
        std::env::var("DATABASE_URL").unwrap_or("not set".to_string())
    );

    crate::core::psql::init().await;
    Ok(())
}

// TODO:
// [] impl concurrency.
#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    init().await?;
    println!("Staring Indexing...");
    index().await;
    Ok(())
}
