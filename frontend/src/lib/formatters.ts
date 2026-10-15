function parseNanoseconds(rt?: number | null): number | null {
  // converts nanoseconds sent from the server to milliseconds
  return rt != null ? rt / 1_000_000 : null;
}

function formatResponseTime(rt?: number | null): string | null {
  // formats nanoseconds sent from the server as a display string
  const ms = parseNanoseconds(rt);
  if (ms === null) return null;
  return ms >= 1000 ? `${(ms / 1000).toFixed(2)} s` : `${ms.toFixed(1)} ms`;
}

function normalizeName(name: string): string {
  return name
    .replace(/^https?:\/\//i, "")
    .replace(/_/g, " ")
    .trim();
}

// Parses Go duration strings (e.g. "12ms", "1.5s", "1m2.3s") to milliseconds.
function parseDurationMs(d: string): number {
  if (!d) return 0;
  let total = 0;
  const re = /(\d+(?:\.\d+)?)(ns|µs|μs|ms|s|m|h)/g;
  let match;
  while ((match = re.exec(d)) !== null) {
    const val = parseFloat(match[1]);
    switch (match[2]) {
      case "ns":
        total += val / 1_000_000;
        break;
      case "µs":
      case "μs":
        total += val / 1_000;
        break;
      case "ms":
        total += val;
        break;
      case "s":
        total += val * 1_000;
        break;
      case "m":
        total += val * 60_000;
        break;
      case "h":
        total += val * 3_600_000;
        break;
    }
  }
  return total;
}

export { parseNanoseconds, formatResponseTime, normalizeName, parseDurationMs }
