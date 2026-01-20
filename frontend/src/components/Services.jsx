import { useState, useMemo } from 'react';
import useServiceData from '../hooks/useServiceData.js';
import ServiceCard from './features/services/ServiceCard.jsx';
import ServiceControls from './features/services/ServiceControls.jsx';

export default function Services() {
    const { data: services, loading, error } = useServiceData();
    const [searchTerm, setSearchTerm] = useState('');
    const [sortKey, setSortKey] = useState('name');

    const sortedAndFilteredServices = useMemo(() => {
        if (!services || services.length === 0) return [];

        const filtered = services.filter(service =>
            service.name.toLowerCase().includes(searchTerm.toLowerCase())
        );

        return [...filtered].sort((a, b) => {
            switch (sortKey) {
                case 'name':
                    return a.name.localeCompare(b.name);
                case 'status':
                    return (a.error === b.error) ? 0 : (a.error ? -1 : 1);
                default:
                    return 0;
            }
        });
    }, [services, searchTerm, sortKey]);

    const renderCards = () => {
        if (error) {
            return <div className="card"><p>Error loading service data: {error}</p></div>;
        }

        if (loading && !services) {
            return <div className="card"><p>Loading services...</p></div>;
        }

        if (!services || services.length === 0) {
            return <div className="card"><p>No Service Data to Display</p></div>;
        }

        if (sortedAndFilteredServices.length === 0) {
            return <div className="card"><p>No services match your search.</p></div>;
        }

        return sortedAndFilteredServices.map((svc, index) => (
            <ServiceCard key={index} service={svc} />
        ));
    };

    return (
        <div className="services-wrapper">
            <h1>Current Services</h1>
            <ServiceControls
                searchTerm={searchTerm}
                setSearchTerm={setSearchTerm}
                sortKey={sortKey}
                setSortKey={setSortKey}
            />
            <div id="cards">
                {renderCards()}
            </div>
        </div>
    );
}
