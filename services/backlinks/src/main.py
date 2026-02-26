import asyncio
import logging
import time
from utils import get_graph_edges, persist_pagerank, NodeMapper
from utils import initialize_connection_pool, close_connection_pool
from page_rank import pagerank
from idf import idf
from dotenv import load_dotenv
import os
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('backlinks.log'),
        logging.StreamHandler()
    ]
)

logger = logging.getLogger(__name__)


def validate_environment():
    """Validate that all required environment variables are set."""
    required_vars = ['PG_HOST', 'PG_PORT', 'PG_DBNAME', 'PG_USER', 'PG_PASSWORD']
    missing_vars = [var for var in required_vars if not os.getenv(var)]
    
    if missing_vars:
        raise ValueError(f"Missing required environment variables: {', '.join(missing_vars)}")
    
    logger.info("Environment variables validated successfully")


async def run_pagerank():
    try:
        start_time = time.time()
        logger.info("Starting PageRank calculation...")
        m = NodeMapper()
        edges = get_graph_edges(m)

        if not edges:
            logger.warning("No edges found in graph. Skipping PageRank calculation.")
            return

        pr = pagerank(edges)
        persist_pagerank(m, pr)
        
        duration = time.time() - start_time
        logger.info(f"PageRank calculation and persistence completed in {duration:.2f}s")
    except Exception as e:
        logger.error(f"PageRank job failed: {e}", exc_info=True)


async def run_idf():
    try:
        start_time = time.time()
        logger.info("Starting IDF calculation...")
        idf()
        
        duration = time.time() - start_time
        logger.info(f"IDF calculation completed in {duration:.2f}s")
    except Exception as e:
        logger.error(f"IDF job failed: {e}", exc_info=True)

async def main():
    load_dotenv("../../.env")
    
    # Validate environment variables
    try:
        validate_environment()
    except ValueError as e:
        logger.error(f"Configuration error: {e}")
        return
    
    # Initialize connection pool
    try:
        initialize_connection_pool(minconn=2, maxconn=10)
    except Exception as e:
        logger.error(f"Failed to initialize connection pool: {e}")
        return
    
    pr_schedule = os.getenv("PR_SCHEDULE", "0 0 * * *")  
    idf_schedule = os.getenv("IDF_SCHEDULE", "0 1 * * *")  
    run_on_startup = os.getenv("BACK_LINKS_RUN_ON_STARTUP", "true").lower() == "true"

    logger.info(f"Configuration: PR_SCHEDULE='{pr_schedule}', IDF_SCHEDULE='{idf_schedule}', RUN_ON_STARTUP={run_on_startup}")

    if run_on_startup:
        logger.info("Running initial calculations on startup...")
        await run_pagerank()
        await run_idf()

    scheduler = AsyncIOScheduler()

    scheduler.add_job(
       run_pagerank,
       trigger=CronTrigger.from_crontab(pr_schedule),
       id='pagerank_job',
       name='PageRank Calculation',
       max_instances=1,
       replace_existing=True
    )

    scheduler.add_job(
        run_idf,
        trigger=CronTrigger.from_crontab(idf_schedule),
        id="idf_job",
        name="IDF Calculation",
        max_instances=1,
        replace_existing=True
    )

    scheduler.start()
    logger.info(f"Scheduler started with PageRank schedule: '{pr_schedule}' and IDF schedule: '{idf_schedule}'")
    
    # Keep running forever
    try:
        while True:
            await asyncio.sleep(3600)
    except asyncio.CancelledError:
        logger.info("Shutting down...")
        scheduler.shutdown()
        close_connection_pool()




if __name__ == "__main__":
    try:
        logger.info("Starting Backlinks Service...")
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Service stopped by user")
    except Exception as e:
        logger.error(f"Service failed with error: {e}", exc_info=True)
    finally:
        close_connection_pool()

