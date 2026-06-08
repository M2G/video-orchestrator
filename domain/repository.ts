import type { VideoJob } from "./job.js";

export interface Repository {
    lockAndMarkProcessing(limit: number): Promise<VideoJob[]>;
    markDone(id: bigint): Promise<void>;
    markRetry(id: bigint, delay: number): Promise<void>;
    markFailed(id: bigint): Promise<void>;
    exists(id: bigint): Promise<boolean>;
}