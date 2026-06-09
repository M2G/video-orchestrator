import type { postgres } from "postgres";
import type { Repository } from "../../domain/repository.js";
import type { VideoJob } from "../../domain/job.js";
import * as db from "../db/generated/query.js";

export function createRepository(sql: postgres.Sql): Repository {
    const instanceID = process.env.INSTANCE_ID ?? (await import("node:os")).hostname();

    return {
        async lockAndMarkProcessing(limit: number): Promise<VideoJob[]> {
            const rows = await db.lockAndMarkProcessing(sql, {
                lockedBy: instanceID,
                limit: BigInt(limit),
            });

            return rows.map((row) => ({
                id: row.id,
                filename: row.filename,
                retryCount: row.retry_count,
            }));
        },

        async markDone(id: bigint): Promise<void> {
            await db.markDone(sql, { id });
        },

        async markRetry(id: bigint, delay: number): Promise<void> {
            await db.markRetry(sql, { id, delay });
        },

        async markFailed(id: bigint): Promise<void> {
            await db.markFailed(sql, { id });
        },

        async exists(id: bigint): Promise<boolean> {
            return db.jobExists(sql, { id });
        },
    };
}