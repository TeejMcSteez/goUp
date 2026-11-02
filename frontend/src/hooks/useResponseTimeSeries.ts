import type { AlignedData } from "uplot";
import { useQueryClient } from "@tanstack/react-query";
import useLiveQuery from "./useLiveQuery";

export interface ResponseTimeSeries {
  names: string[];
  seq: number;
  // [unixSeconds, service1Ms, service2Ms, ...] with nulls where a service
  // has no sample at a given time
  data: AlignedData;
}

// Mirrors server/protocol.go encodeFrame. All integers little-endian.
//
//   0  u8   frame type (0 = snapshot, 1 = append)
//   1  u8   padding
//   2  u16  column count
//   4  u32  seq
//   8  u32  row count
//  12  name table, snapshot only: per column [u8 len][utf8 bytes],
//      padded to a 4 byte boundary
//      ts:   rows x u32 unix seconds (ascending)
//      cols: columns x rows x u32 response time in microseconds, 0 = no sample
const FRAME_SNAPSHOT = 0;
const FRAME_APPEND = 1;
const HEADER_SIZE = 12;
const US_PER_MS = 1_000;

const align4 = (n: number) => (n + 3) & ~3;

const LITTLE_ENDIAN = new Uint8Array(new Uint32Array([1]).buffer)[0] === 1;

// Zero-copy view on little-endian hosts (every browser in practice), with a
// DataView fallback so a big-endian host still decodes correctly.
function u32View(buf: ArrayBuffer, offset: number, length: number) {
  if (LITTLE_ENDIAN) return new Uint32Array(buf, offset, length);
  const dv = new DataView(buf, offset, length * 4);
  const out = new Uint32Array(length);
  for (let i = 0; i < length; i++) out[i] = dv.getUint32(i * 4, true);
  return out;
}

export function decodeFrame(buf: ArrayBuffer) {
  if (buf.byteLength < HEADER_SIZE)
    throw new Error("Response time frame too short");

  const dv = new DataView(buf);
  const typ = dv.getUint8(0);
  const colCount = dv.getUint16(2, true);
  const seq = dv.getUint32(4, true);
  const rows = dv.getUint32(8, true);

  if (typ !== FRAME_SNAPSHOT && typ !== FRAME_APPEND)
    throw new Error(`Unknown response time frame type ${typ}`);

  const names: string[] = [];
  let o = HEADER_SIZE;
  if (typ === FRAME_SNAPSHOT) {
    const bytes = new Uint8Array(buf);
    const decoder = new TextDecoder();
    for (let c = 0; c < colCount; c++) {
      const len = bytes[o++];
      names.push(decoder.decode(bytes.subarray(o, o + len)));
      o += len;
    }
  }
  o = align4(o);

  if (o + rows * 4 * (1 + colCount) > buf.byteLength)
    throw new Error("Response time frame truncated");

  const ts = u32View(buf, o, rows);
  o += rows * 4;

  const cols: (number | null)[][] = [];
  for (let c = 0; c < colCount; c++) {
    const raw = u32View(buf, o, rows);
    o += rows * 4;
    const ms: Array<number | null> = Array.from({ length: rows });
    for (let i = 0; i < rows; i++) {
      const v = raw[i];
      ms[i] = v === 0 ? null : v / US_PER_MS;
    }
    cols.push(ms);
  }

  return { typ, seq, names, ts, cols };
}

type Frame = ReturnType<typeof decodeFrame>;
type SeriesColumns = [ArrayLike<number>, ...(number | null)[][]];

// Nested under "responseTime" so websocket pushes invalidate it too.
const QUERY_KEY = ["responseTime", "series"] as const;

// GET /api/rt: the full history as a snapshot frame.
async function fetchSnapshot(): Promise<ResponseTimeSeries | null> {
  const res = await fetch("/api/rt");
  if (!res.ok)
    throw new Error(`Error fetching response times: ${res.statusText}`);

  const frame = decodeFrame(await res.arrayBuffer());
  if (frame.typ !== FRAME_SNAPSHOT)
    throw new Error("Expected a snapshot frame from /api/rt");
  if (frame.ts.length === 0 || frame.names.length === 0) return null;

  return {
    names: frame.names,
    seq: frame.seq,
    data: [frame.ts, ...frame.cols] as AlignedData,
  };
}

// Maps each frame column to its column index in prev. Frames carrying a name
// table are matched by name, since a window only holds the services that
// reported in it. Nameless append frames follow the snapshot's column order.
// Returns null when the frame can't be lined up (e.g. a new service), so the
// caller falls back to a fresh snapshot.
function mapColumns(prev: ResponseTimeSeries, frame: Frame): number[] | null {
  if (frame.typ === FRAME_SNAPSHOT) {
    const index = new Map(prev.names.map((name, i) => [name, i]));
    const map = frame.names.map((name) => index.get(name) ?? -1);
    return map.includes(-1) ? null : map;
  }
  if (frame.cols.length !== prev.names.length) return null;
  return frame.cols.map((_, i) => i);
}

// Appends the frame's rows onto prev without mutating it. The window starts
// at prev's last timestamp (inclusive), so a row at that same second merges
// into prev's last row instead of duplicating it.
function mergeFrame(
  prev: ResponseTimeSeries,
  frame: Frame,
  colMap: number[],
): ResponseTimeSeries {
  const [prevTs, ...prevCols] = prev.data as unknown as SeriesColumns;
  const prevLen = prevTs.length;
  const lastTs = prevTs[prevLen - 1];

  let start = 0;
  while (start < frame.ts.length && frame.ts[start] <= lastTs) start++;
  const added = frame.ts.length - start;

  const ts = new Uint32Array(prevLen + added);
  ts.set(prevTs);
  ts.set(frame.ts.subarray(start), prevLen);

  const pad: null[] = Array.from({ length: added }, () => null);
  const cols = prevCols.map((col) => col.concat(pad));

  frame.cols.forEach((col, c) => {
    const target = cols[colMap[c]];
    for (let i = 0; i < col.length; i++) {
      const v = col[i];
      if (v == null || frame.ts[i] < lastTs) continue;
      target[i < start ? prevLen - 1 : prevLen + i - start] = v;
    }
  });

  return {
    // Same reference keeps chart options (built from names) stable.
    names: prev.names,
    seq: frame.seq,
    data: [ts, ...cols] as AlignedData,
  };
}

// POST /api/rt: only rows recorded since prev's newest timestamp.
async function fetchAppend(
  prev: ResponseTimeSeries,
): Promise<ResponseTimeSeries | null> {
  const prevTs = prev.data[0];
  const lastTs = prevTs[prevTs.length - 1];

  const res = await fetch("/api/rt", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    // Unix seconds, same unit as the frame; the server fills `to` with its
    // own clock so browser clock skew can't hide fresh rows.
    body: JSON.stringify({ from: lastTs }),
  });
  if (!res.ok)
    throw new Error(`Error fetching response time update: ${res.statusText}`);

  const frame = decodeFrame(await res.arrayBuffer());
  if (frame.ts.length === 0) return prev;

  const colMap = mapColumns(prev, frame);
  if (!colMap) return fetchSnapshot();
  return mergeFrame(prev, frame, colMap);
}

export default function useResponseTimeSeries() {
  const queryClient = useQueryClient();

  // First fetch (or after the cache was dropped) pulls the full snapshot;
  // every refetch after that only pulls the rows newer than what's cached.
  const fetchSeries = () => {
    const prev = queryClient.getQueryData<ResponseTimeSeries | null>(
      QUERY_KEY,
    );
    return prev ? fetchAppend(prev) : fetchSnapshot();
  };

  return useLiveQuery(QUERY_KEY, fetchSeries);
}
