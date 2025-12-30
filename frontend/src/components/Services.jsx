import { useState, useEffect } from 'react';

function Card({ service }) {
    const { name, response, response_time, data, error } = service;

    return (
        <div className="card">
            <h3 className="svcName">Name: <a href={name} target="_blank" rel="noopener noreferrer">{name}</a></h3>
            <p>Status: {error ? "❌ Error" : "✅ Operational"}</p>
            <p>Response Time: {response_time}</p>
            <p className="svcHttpRes">HTTP Response: {response}</p>
            <h2>API Response</h2>
            <div className="svcData">{data ? data : "No API setup in configuration"}</div>
        </div>
    );
}

export default function Services() {
    const [services, setServices] = useState([]);
    const [error, setError] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');

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

    const filteredServices = services.filter(service =>
        service.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const renderCards = () => {
        if (error) {
            return <div className="card"><p>Error loading service data: {error}</p></div>;
        }

        if (services.length === 0) {
            return <div className="card"><p>No Service Data to Display</p></div>;
        }

        if (filteredServices.length === 0) {
            return <div className="card"><p>No services match your search.</p></div>;
        }

        return filteredServices.map((svc, index) => (
            <Card key={index} service={svc} />
        ));
    };

    return (
        <div className="services-wrapper">
            <h1>Current Services</h1>
            <input
                type="text"
                placeholder="Search services..."
                className="search-bar"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
            />
            <div id="cards">
                {renderCards()}
            </div>
        </div>
    );
}
