import { createPool } from "@infrastructure/db/pool.js";
import { createRepository } from "@infrastructure/repository/pg_repository.js";
import { createLogger } from "@infrastructure/logger/logger.js";
import { createOrchestrator } from "@application/orchestrator.js";
import { createVideoHandler } from "@interfaces/handler.js";
import { createScheduler } from "@interfaces/scheduler.js";
import { createWatcher } from "@interfaces/watcher.js";

async function main(): Promise<void> {
    const log = createLogger(
        (process.env.LOG_LEVEL as "debug" | "info" | "warn" | "error") ?? "info",
    );

    const connString = process.env.DATABASE_URL;
    if (!connString) {
        log.error({}, "DATABASE_URL is not set");
        process.exit(1);
    }

    const doneDir = process.env.DONE_DIR;
    if (!doneDir) {
        log.error({}, "DONE_DIR is not set");
        process.exit(1);
    }

    const outputDir = process.env.OUTPUT_DIR;
    if (!outputDir) {
        log.error({}, "OUTPUT_DIR is not set");
        process.exit(1);
    }

    const workers = parseInt(process.env.WORKERS ?? "4", 10);
    const maxRetries = parseInt(process.env.MAX_RETRIES ?? "5", 10);
    const intervalMs = parseInt(process.env.SCHEDULER_INTERVAL_MS ?? "5000", 10);

    const sql = createPool(connString);
    const repo = createRepository(sql);
    const handler = createVideoHandler(outputDir, log);
    const orch = createOrchestrator(repo, handler, workers, maxRetries);
    const scheduler = createScheduler(orch, intervalMs, log);
    const watcher = createWatcher(doneDir, log, repo);

    scheduler.start();
    watcher.start();

    log.info({}, "video_orchestrator_started");

    async function shutdown(): Promise<void> {
        log.info({}, "shutting_down");
        scheduler.stop();
        watcher.stop();
        await sql.end();
        log.info({}, "shutdown_complete");
        process.exit(0);
    }

    process.on("SIGTERM", shutdown);
    process.on("SIGINT", shutdown);
}

main().catch((err) => {
    console.error(err);
    process.exit(1);
});