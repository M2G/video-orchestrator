export function nextDelay(retry: number): number {
    const base = 5;
    const max = 300;

    const delay = Math.min(base * (1 << retry), max);
    const jitter = Math.floor(Math.random() * 5);

    return delay + jitter;
}