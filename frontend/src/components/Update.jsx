import { useState, useEffect } from "react";

export default function Update() {
  const [schedule, setSchedule] = useState(null);
  const [timespan, setTimespan] = useState(30);
  const [interval, setInterval] = useState("seconds");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchSchedule = async () => {
    try {
      setLoading(true);
      const res = await fetch("/api/schedule");
      if (!res.ok) {
        throw new Error("Could not get schedule data");
      }
      const json = await res.json();
      setSchedule(json);
      setTimespan(json.timespan);
      setInterval(json.interval);
      setError(null);
    } catch (err) {
      setError(err.message);
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
        body: JSON.stringify({
          timespan: Number(timespan),
          interval: interval,
        }),
      });

      if (!req.ok) {
        alert("Update failed!");
        return;
      }
      // Refresh the schedule data to show the update
      await fetchSchedule();
    } catch (err) {
      alert(`Update failed: ${err.message}`);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center">
        <p>Loading schedule...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center">
        <h3>Failed to get schedule!</h3>
        <p>{error}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center">
      {schedule && (
        <p>
          Current: {schedule.timespan} {schedule.interval}
        </p>
      )}
      <form className="flex flex-row p-4 m-4 items-center justify-center text-center gap-4 flex-wrap">
        <label htmlFor="timespan">Timespan: {timespan}</label>
        <input
          id="timespan"
          type="range"
          min="1"
          max="60"
          value={timespan}
          onChange={(e) => setTimespan(e.target.value)}
          required
        />
        <label htmlFor="selectTimespan">interval</label>
        <select
          name="timespans"
          id="selectTimespan"
          value={interval}
          onChange={(e) => setInterval(e.target.value)}
          className="flex items-center justify-center text-center rounded-lg p-2"
        >
          <option value="seconds">Seconds</option>
          <option value="minutes">Minutes</option>
          <option value="hours">Hours</option>
        </select>
      </form>
      <button
        className="flex items-center justify-center gap-4 w-full max-w-[600px] mx-auto text-center rounded-2xl p-4 hover:-translate-y-[5px] hover:border-primary"
        onClick={handleUpdateSchedule}
      >
        Update
      </button>
    </div>
  );
}
