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
}

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub batch_size: usize,
    pub loop_delay_ms: u64,
}

pub fn load_config(env_path: String) -> Config {
    from_path(env_path).expect("cannot open .env file");
    let app = load_app_config();
    let psql = load_psql_config();
    Config { app, psql }
}

fn load_app_config() -> AppConfig {
    let batch_size = env::var("BATCH_SIZE")
        .unwrap_or_else(|_| "100".to_string())
        .parse::<usize>()
        .expect("BATCH_SIZE must be a number");
    let loop_delay_ms = env::var("LOOP_DELAY_MS")
        .unwrap_or_else(|_| "100".to_string())
        .parse::<u64>()
        .expect("LOOP_DELAY_MS must be a number");
    AppConfig {
        batch_size,
        loop_delay_ms,
    }
}

fn load_psql_config() -> PsqlConfig {
    let host = env::var("PG_HOST").expect("PG_HOST must be set");
    let port = env::var("PG_PORT").expect("PG_PORT must be set");
    let user = env::var("PG_USER").expect("PG_USER must be set");
    let password = env::var("PG_PASSWORD").expect("PG_PASSWORD must be set");
    let dbname = env::var("PG_DBNAME").expect("PG_DBNAME must be set");

    let url = format!("postgres://{user}:{password}@{host}:{port}/{dbname}");
    PsqlConfig { url }
}
