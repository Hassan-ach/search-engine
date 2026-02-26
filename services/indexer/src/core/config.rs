use dotenv::from_path;
use std::env;

#[derive(Debug, Clone)]
pub struct Config {
    pub app: AppConfig,
    pub psql: PsqlConfig,
}

#[derive(Debug, Clone)]
pub struct PsqlConfig {
    pub url: String,
    pub max_connections: u32,
    pub min_connections: u32,
    pub acquire_timeout_seconds: std::time::Duration,
}

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub log_path: String,
    pub loop_delay_ms: u64,
}

pub fn load_config(env_path: String) -> Config {
    from_path(env_path).expect("cannot open .env file");
    let app = load_app_config();
    let psql = load_psql_config();
    Config { app, psql }
}

fn load_app_config() -> AppConfig {
    let log_path = env::var("LOG_PATH").unwrap_or_else(|_| "indexer.log".to_string());
    let loop_delay_ms = env::var("LOOP_DELAY_MS")
        .unwrap_or_else(|_| "100".to_string())
        .parse::<u64>()
        .expect("LOOP_DELAY_MS must be a number");

    AppConfig {
        log_path,
        loop_delay_ms,
    }
}

fn load_psql_config() -> PsqlConfig {
    let host = env::var("PG_HOST").expect("PG_HOST must be set");
    let port = env::var("PG_PORT").expect("PG_PORT must be set");
    let user = env::var("PG_USER").expect("PG_USER must be set");
    let password = env::var("PG_PASSWORD").expect("PG_PASSWORD must be set");
    let dbname = env::var("PG_DBNAME").expect("PG_DBNAME must be set");
    let max_connections = env::var("PG_MAX_CONNECTIONS")
        .unwrap_or_else(|_| "10".to_string())
        .parse::<u32>()
        .expect("PG_MAX_CONNECTIONS must be a number");
    let min_connections = env::var("PG_MIN_CONNECTIONS")
        .unwrap_or_else(|_| "2".to_string())
        .parse::<u32>()
        .expect("PG_MIN_CONNECTIONS must be a number");
    let acquire_timeout = env::var("ACQUIRE_TIMEOUT_SECONDS")
        .unwrap_or_else(|_| "5".to_string())
        .parse::<u64>()
        .expect("ACQUIRE_TIMEOUT_SECONDS must be a number");

    let url = format!("postgres://{user}:{password}@{host}:{port}/{dbname}");
    PsqlConfig {
        url,
        max_connections,
        min_connections,
        acquire_timeout_seconds: std::time::Duration::from_secs(acquire_timeout),
    }
}
