/** Simple in-flight concurrency limiter with AbortSignal support. */
export class Semaphore {
  private active = 0;
  private readonly waiters: Array<{
    resolve: () => void;
    reject: (err: Error) => void;
    signal?: AbortSignal;
    onAbort?: () => void;
  }> = [];

  constructor(private readonly max: number) {}

  async acquire(signal?: AbortSignal): Promise<() => void> {
    if (signal?.aborted) {
      throw signal.reason instanceof Error ? signal.reason : new Error("Aborted");
    }
    if (this.active < this.max) {
      this.active++;
      return () => this.release();
    }
    await new Promise<void>((resolve, reject) => {
      const entry: (typeof this.waiters)[number] = { resolve, reject, signal };
      entry.onAbort = () => {
        const idx = this.waiters.indexOf(entry);
        if (idx >= 0) {
          this.waiters.splice(idx, 1);
        }
        reject(signal!.reason instanceof Error ? signal!.reason : new Error("Aborted"));
      };
      if (signal) {
        signal.addEventListener("abort", entry.onAbort, { once: true });
      }
      this.waiters.push(entry);
    });
    this.active++;
    return () => this.release();
  }

  private release(): void {
    this.active = Math.max(0, this.active - 1);
    const next = this.waiters.shift();
    if (next) {
      if (next.signal && next.onAbort) {
        next.signal.removeEventListener("abort", next.onAbort);
      }
      next.resolve();
    }
  }
}
