import { useState, useEffect } from "react";

export default function Update() {
    const [schedule, setSchedule] = useState(null);
    const [timespan, setTimespan] = useState(30);
    const [timeInterval, setTimeInterval] = useState("seconds");
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
            setTimeInterval(json.timeInterval);
            setError(null);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSchedule();
    }, []);

    const handleUpdateSchedule = async () => {
        try {
            const req = await fetch("/api/schedule", {
                method: "POST",
                body: JSON.stringify({
                    timespan: Number(timespan),
                    timeInterval: timeInterval,
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
            <div>
                <h2>Update Schedule</h2>
                <div id="schedule">
                    <p>Loading schedule...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div>
                <h2>Update Schedule</h2>
                <div id="schedule">
                    <h3>Failed to get schedule!</h3>
                    <p>{error}</p>
                </div>
            </div>
        );
    }

    return (
        <div>
            <h2>Update Schedule</h2>
            <div id="schedule">
                {schedule && (
                    <p>Current: {schedule.timespan} {schedule.timeInterval}</p>
                )}
                <form id="scheduleForm">
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
                    <label htmlFor="selectTimespan">Interval</label>
                    <select 
                        name="timespans" 
                        id="selectTimespan" 
                        value={timeInterval}
                        onChange={(e) => setTimeInterval(e.target.value)}
                    >
                        <option value="seconds">Seconds</option>
                        <option value="minutes">Minutes</option>
                        <option value="hours">Hours</option>
                    </select>
                </form>
                <button onClick={handleUpdateSchedule}>Update</button>
            </div>
        </div>
    );
}
