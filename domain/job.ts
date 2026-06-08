export interface VideoJob {
    id: bigint;
    filename: string;
    retryCount: number;
}

export function canRetry(job: VideoJob, max: number): boolean {
    return job.retryCount < max;
}