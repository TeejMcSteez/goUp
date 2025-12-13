import { useState, useEffect } from 'react';

function Card({ service }) {
    const { name, response, response_time, data } = service;
    const isSuccess = response === "200";

    return (
        <div className="card">
            <h1 className="svcName">Server URL: <a href={name} target="_blank" rel="noopener noreferrer">{name}</a></h1>
            <p>Status: {isSuccess ? "✅" : "❌"} in {response_time}</p>
            <p className="svcHttpRes">HTTP Response: {response}</p>
            <h2>API Response</h2>
            <div className="svcData">{data ? data : "No API setup in configuration"}</div>
        </div>
    );
}

export default function Services() {
    const [services, setServices] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        const getServiceData = async () => {
            try {
                const res = await fetch("/api");
                if (!res.ok) {
                    throw new Error(`Server error: ${res.status}`);
                }
                const data = await res.json();
                if (Array.isArray(data)) {
                    setServices(data);
                } else {
                    console.error("Expected array from /api, got:", data);
                    setServices([]);
                }
            } catch (err) {
                setError(err.message);
                console.error("Error fetching service data:", err);
            }
        };

        getServiceData();
        
        const intervalId = setInterval(getServiceData, 5000); // Refresh every 5 seconds

        return () => clearInterval(intervalId); // Cleanup on component unmount
    }, []);

    if (error) {
        return <div id="cards"><p>Error loading service data: {error}</p></div>;
    }

    if (services.length === 0) {
        return <div id="cards"><div className="card"><p>No Service Data to Display</p></div></div>;
    }

    return (
        <>
            <h1>Current Services</h1>
            <div id="cards">
                {services.map((svc, index) => (
                    <Card key={index} service={svc} />
                ))}
            </div>
        </>
    );
}
