mod core;

use crate::core::config::load_config;
use crate::core::indexer::Indexe;
use crate::core::indexer::Indexer;
use crate::core::psql::Psql;
use std::error::Error;
use std::sync::Arc;
use tokio::task::JoinHandle;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let conf = load_config("../../.env".to_string());
    let psql = Psql::new(conf.psql).await?;
    let indx = Arc::new(Indexer::<Psql>::new(psql, conf.app));

    let token = tokio_util::sync::CancellationToken::new();

    let mut handles = Vec::<JoinHandle<()>>::new();

    for i in 0..4 {
        println!("Starting indexer task... id: {}", i);
        let token_clone = token.clone();
        let job = indx.clone();
        handles.push(tokio::spawn(async move {
            job.start(token_clone).await;
        }));
    }

    handles.push(tokio::spawn(async move {
        match tokio::signal::ctrl_c().await {
            Ok(()) => {
                println!("Ctrl-C received, sending shutdown signal...");
                token.cancel();
            }
            Err(err) => {
                println!("Unable to listen for shutdown signal: {}", err);
            }
        }
    }));

    for handle in handles {
        if let Err(err) = handle.await {
            println!("Indexer task failed: {}", err);
        }
    }

    Ok(())
}
