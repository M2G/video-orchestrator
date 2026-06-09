type Level = "debug" | "info" | "warn" | "error";

const LEVELS: Record<Level, number> = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
};

export interface Logger {
    debug(obj: object, msg: string): void;
    info(obj: object, msg: string): void;
    warn(obj: object | string, msg?: string): void;
    error(obj: object, msg: string): void;
}

export function createLogger(minLevel: Level = "info"): Logger {
    function log(level: Level, obj: object, msg: string): void {
        if (LEVELS[level] < LEVELS[minLevel]) return;

        process.stdout.write(
            JSON.stringify({
                level,
                time: new Date().toISOString(),
                ...obj,
                msg,
            }) + "\n",
        );
    }

    return {
        debug: (obj, msg) => log("debug", obj, msg),
        info: (obj, msg) => log("info", obj, msg),
        warn: (obj, msg) => {
            if (typeof obj === "string") {
                log("warn", {}, obj);
            } else {
                log("warn", obj, msg ?? "");
            }
        },
        error: (obj, msg) => log("error", obj, msg),
    };
}