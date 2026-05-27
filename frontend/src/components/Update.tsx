import { useState, useEffect } from "react";
import type { Schedule } from "../types";

export default function Update() {
  const [schedule, setSchedule] = useState<Schedule | null>(null);
  const [timespan, setTimespan] = useState(30);
  const [intervalUnit, setIntervalUnit] = useState("seconds");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchSchedule = async () => {
    try {
      setLoading(true);
      const res = await fetch("/api/schedule");
      if (!res.ok) throw new Error("Could not get schedule data");
      const json: Schedule = await res.json();
      setSchedule(json);
      setTimespan(json.timespan);
      setIntervalUnit(json.interval);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const id = setTimeout(fetchSchedule, 0);
    return () => clearTimeout(id);
  }, []);

  const handleUpdateSchedule = async () => {
    try {
      const req = await fetch("/api/schedule", {
        method: "POST",
        body: JSON.stringify({ timespan: Number(timespan), interval: intervalUnit }),
      });
      if (!req.ok) {
        alert("Update failed!");
        return;
      }
      await fetchSchedule();
    } catch (err) {
      alert(`Update failed: ${(err as Error).message}`);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-6">
        <div className="w-5 h-5 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 p-6">
        <p className="text-error text-sm font-medium">Failed to load schedule</p>
        <p className="text-muted text-xs">{error}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-4">
      {schedule && (
        <p className="text-sm text-muted text-center">
          Current: <span className="text-fg font-medium">{schedule.timespan} {schedule.interval}</span>
        </p>
      )}

      <div className="flex flex-col sm:flex-row items-center gap-4 w-full">
        <div className="flex flex-col gap-1 w-full sm:flex-1">
          <label htmlFor="timespan" className="text-xs text-muted">
            Timespan: <span className="text-fg font-medium">{timespan}</span>
          </label>
          <input
            id="timespan"
            type="range"
            min="1"
            max="60"
            value={timespan}
            onChange={(e) => setTimespan(Number(e.target.value))}
            required
            className="w-full accent-primary"
          />
        </div>

        <div className="flex flex-col gap-1 w-full sm:w-auto">
          <label htmlFor="selectTimespan" className="text-xs text-muted">
            Interval
          </label>
          <select
            name="timespans"
            id="selectTimespan"
            value={intervalUnit}
            onChange={(e) => setIntervalUnit(e.target.value)}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm text-fg w-full sm:w-auto"
          >
            <option value="seconds">Seconds</option>
            <option value="minutes">Minutes</option>
            <option value="hours">Hours</option>
          </select>
        </div>
      </div>

      <button
        onClick={handleUpdateSchedule}
        className="w-full rounded-lg border border-border px-4 py-2 text-sm font-medium text-fg transition-colors hover:border-primary hover:text-primary"
      >
        Update Schedule
      </button>
    </div>
  );
}
