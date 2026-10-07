import uPlot, { type AlignedData } from "uplot";
import useLiveQuery from "./useLiveQuery";
import { parseDurationMs } from "./useResponseTimeData.ts";
import type { ResponseTimeEntry } from "../types";

export interface ResponseTimeSeries {
  names: string[];
  // [unixSeconds, service1Ms, service2Ms, ...] with nulls where a service
  // has no sample at a given time
  data: AlignedData;
}

async function fetchSeries(): Promise<ResponseTimeSeries | null> {
  const res = await fetch("/api/rt");
  if (!res.ok)
    throw new Error(`Error fetching response times: ${res.statusText}`);

  const json: unknown = await res.json();
  if (!Array.isArray(json) || json.length === 0) return null;

  const items = json as ResponseTimeEntry[];

  const byService = new Map<string, { t: number; ms: number }[]>();
  for (const item of items) {
    const ms = parseDurationMs(item.response_time);
    const t = Date.parse(item.service_data.timestamp) / 1000;
    if (ms <= 0 || Number.isNaN(t)) continue;
    const name = item.service_data.name;
    const points = byService.get(name) ?? [];
    points.push({ t, ms });
    byService.set(name, points);
  }

  if (byService.size === 0) return null;

  const names = [...byService.keys()];
  const tables = names.map((name) => {
    const points = byService.get(name)!.sort((a, b) => a.t - b.t);
    return [points.map((p) => p.t), points.map((p) => p.ms)] as AlignedData;
  });

  return { names, data: uPlot.join(tables) };
}

export default function useResponseTimeSeries() {
  // Nested under "responseTime" so websocket pushes invalidate it too.
  return useLiveQuery(["responseTime", "series"], fetchSeries);
}
