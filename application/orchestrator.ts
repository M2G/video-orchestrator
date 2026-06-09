import type { Repository } from "../domain/repository.js";
import type { VideoJob } from "../domain/job.js";
import { canRetry } from "../domain/job.js";
import { nextDelay } from "../domain/retry.js";
import { createCircuitBreaker, type CircuitBreaker } from "./circuit_breaker.js";
import type { Logger } from "../infrastructure/logger";

export interface Handler {
    handle(job: VideoJob): Promise<void>;
}

export interface Orchestrator {
    runOnce(log: Logger): Promise<void>;
}

async function handleFailure(
    repo: Repository,
    job: VideoJob,
    maxRetries: number,
    log: Logger,
): Promise<void> {
    if (!canRetry(job, maxRetries)) {
        try {
            await repo.markFailed(job.id);
        } catch (err) {
            log.error({ err, job_id: job.id }, "mark_failed_error");
        }
        return;
    }

    const delay = nextDelay(job.retryCount);

    log.warn({ job_id: job.id, retry: job.retryCount, delay }, "job_retry");

    try {
        await repo.markRetry(job.id, delay);
    } catch (err) {
        log.error({ err, job_id: job.id }, "mark_retry_error");
    }
}

export function createOrchestrator(
    repo: Repository,
    handler: Handler,
    workers: number,
    maxRetries: number,
): Orchestrator {
    const breaker: CircuitBreaker = createCircuitBreaker(5, 10_000);

    return {
        async runOnce(log: Logger): Promise<void> {
            let jobs: VideoJob[];

            try {
                jobs = await repo.lockAndMarkProcessing(workers);
            } catch (err) {
                log.error({ err }, "lock_failed");
                return;
            }

            log.info({ jobs_count: jobs.length }, "jobs_locked");

            await Promise.all(
                jobs.map(async (job) => {
                    log.info({ job_id: job.id }, "job_started");
                    const start = Date.now();

                    if (!breaker.allow()) {
                        log.warn("circuit_breaker_open");
                        await handleFailure(repo, job, maxRetries, log);
                        return;
                    }

                    try {
                        await handler.handle(job);
                    } catch (err) {
                        log.error({ err, job_id: job.id }, "job_failed");
                        breaker.fail();
                        await handleFailure(repo, job, maxRetries, log);
                        return;
                    }

                    breaker.success();

                    try {
                        await repo.markDone(job.id);
                    } catch (err) {
                        log.error({ err, job_id: job.id }, "mark_done_error");
                    }

                    log.info({ job_id: job.id, duration_ms: Date.now() - start }, "job_done");
                }),
            );
        },
    };
}