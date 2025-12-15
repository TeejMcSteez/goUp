import { useState, useEffect } from "react";

export default function QuickServices() {
    const [qs, setQs] = useState([])
    const [error, setError] = useState(null);

    useEffect(() => {
        const getDownServices = async () => {
            try {
                const res = await fetch("/api/status");
                if (!res.ok) {
                    throw new Error(`Server error: ${res.status}`);
                }
                const data = await res.json();
                if (Array.isArray(data.downed_services)) {
                    setQs(data.downed_services)
                } else {
                    console.error(`Expected array from /api/status, got: ${data}`)
                }
            } catch (err) {
                setError(err.message);
                console.error("Error fetching quick service data:", err);
            }
        };

        getDownServices();

        const intervalId = setInterval(getDownServices, 5000);

        return () => clearInterval(intervalId);
    }, []);

    if (error) {
        return <div id="quickServices">Error Loading quick service data: {error}</div>
    }

    if (qs.length === 0) {
        return <div id="quickServices">
            <a href="#cards" id="quickServicesCard">All Systems Operational ✅</a>
        </div>
    }
    let numberOfDownedServices = qs.length;
    return(
        <div id="quickServices">
            <a href="#cards" id="quickServicesCard">{numberOfDownedServices} Errors Detected ❌</a>
        </div>
    );
}