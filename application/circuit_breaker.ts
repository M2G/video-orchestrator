type State = "closed" | "open" | "half-open";

export interface CircuitBreaker {
    allow(): boolean;
    fail(): void;
    success(): void;
}

export function createCircuitBreaker(
    threshold: number,
    resetAfter: number,
): CircuitBreaker {
    let state: State = "closed";
    let failures = 0;
    let lastFailure = 0;

    return {
        allow() {
            switch (state) {
                case "closed":
                    return true;
                case "open":
                    if (Date.now() - lastFailure > resetAfter) {
                        state = "half-open";
                        return true;
                    }
                    return false;
                case "half-open":
                    return false;
            }
        },

        fail() {
            failures++;
            lastFailure = Date.now();
            if (failures >= threshold) {
                state = "open";
            }
        },

        success() {
            failures = 0;
            state = "closed";
        },
    };
}