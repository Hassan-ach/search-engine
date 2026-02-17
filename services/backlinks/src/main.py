import asyncio
from  utils import get_graph_edges
from page_rank import pagerank
from  utils import persist_pagerank
from  idf import idf
from utils import NodeMapper
from dotenv import load_dotenv
import os
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

async def run_pagerank():
    print("Running PageRank calculation...")
    m = NodeMapper()
    edges = get_graph_edges(m)

    if not edges:
        return

    pr = pagerank(edges)

    persist_pagerank(m, pr)
    print("PageRank calculation and persistence completed.")


async def run_idf():
    print("Running IDF calculation...")
    idf()
    print("IDF calculation completed.")

async def main():
    load_dotenv("../../.env")
    pr_schedule = os.getenv("PR_SCHEDULE", "0 0 * * *")  
    idf_schedule = os.getenv("IDF_SCHEDULE", "0 1 * * *")  
    run_on_startup = os.getenv("BACK_LINKS_RUN_ON_STARTUP", "true").lower() == "true"

    if run_on_startup:
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
    print("Scheduler started with PageRank schedule: '{}' and IDF schedule: '{}'".format(pr_schedule, idf_schedule))
    
    # Keep running forever
    try:
        while True:
            await asyncio.sleep(3600)
    except asyncio.CancelledError:
        print("Shutting down...")
        scheduler.shutdown()




if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("Service stopped")

