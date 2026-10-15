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

export { parseNanoseconds, formatResponseTime }
