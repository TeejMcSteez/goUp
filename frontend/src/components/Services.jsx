import { useState, useEffect } from 'react';

function Card({ service }) {
    const { name, response, response_time, data, error } = service;
    const [showApiResponse, setShowApiResponse] = useState(false);

    const toggleApiResponse = () => {
        setShowApiResponse(!showApiResponse);
    };

    return (
        <div className="card">
            <h3 className="svcName">Name: {name}</h3>
            <p>Status: {error ? "❌ Error" : "✅ Operational"}</p>
            <p>Response Time: {response_time}</p>
            <p className="svcHttpRes">HTTP Response: {response}</p>
            <h2 onClick={toggleApiResponse} style={{ cursor: 'pointer' }}>
                API Response {showApiResponse ? '▲' : '▼'}
            </h2>
            {showApiResponse && (
                <div className="svcData">{data ? data : "No API setup in configuration"}</div>
            )}
        </div>
    );
}

export default function Services() {
    const [services, setServices] = useState([]);
    const [error, setError] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [sortKey, setSortKey] = useState('name'); // 'name', 'status'

    useEffect(() => {
        const getServiceData = async () => {
            try {
                const res = await fetch("/api");
                if (!res.ok) {
                    throw new Error(`Server error: ${res.status}`);
                }
                const data = await res.json();
                if (data && Array.isArray(data)) {
                    setServices(data);
                } else {
                    console.error("Expected object with a 'services' array from /api, got:", data);
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

    const sortedAndFilteredServices = [...filteredServices].sort((a, b) => {
        let compare = 0;

        switch (sortKey) {
            case 'name':
                compare = a.name.localeCompare(b.name);
                break;
            case 'status':
                // Error (true) comes before Operational (false)
                compare = (a.error === b.error) ? 0 : (a.error ? -1 : 1);
                break;
            default:
                compare = 0;
        }
        return compare;
    });

    const renderCards = () => {
        if (error) {
            return <div className="card"><p>Error loading service data: {error}</p></div>;
        }

        if (services.length === 0) {
            return <div className="card"><p>No Service Data to Display</p></div>;
        }

        if (sortedAndFilteredServices.length === 0) {
            return <div className="card"><p>No services match your search.</p></div>;
        }

        return sortedAndFilteredServices.map((svc, index) => (
            <Card key={index} service={svc} />
        ));
    };

    return (
        <div className="services-wrapper">
            <h1>Current Services</h1>
            <div className="controls">
                <input
                    type="text"
                    placeholder="Search services..."
                    className="search-bar"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
                <select value={sortKey} onChange={(e) => setSortKey(e.target.value)} className="sort-select">
                    <option value="name">Sort by Name</option>
                    <option value="status">Sort by Status</option>
                </select>
            </div>
            <div id="cards">
                {renderCards()}
            </div>
        </div>
    );
}
