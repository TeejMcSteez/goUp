import { useState, useEffect } from 'react';
import Island from '../layout/Island.jsx';
import QuickServices from '../QuickServices.jsx';

function QuickStats() {
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
                if (data && Array.isArray(data)) {
                    setServices(data);
                } else {
                    setServices([]);
                }
            } catch (err) {
                setError(err.message);
            }
        };

        getServiceData();
        const intervalId = setInterval(getServiceData, 5000);
        return () => clearInterval(intervalId);
    }, []);

    const totalServices = services.length;
    const errorServices = services.filter(s => s.error).length;
    const operationalServices = totalServices - errorServices;
    const uptimePercentage = totalServices > 0
        ? ((operationalServices / totalServices) * 100).toFixed(1)
        : 0;

    if (error) {
        return <p>Error loading stats: {error}</p>;
    }

    return (
        <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: 'var(--space-lg)',
            marginBottom: 'var(--space-xl)'
        }}>
            <div style={{
                textAlign: 'center',
                padding: 'var(--space-lg)',
                background: 'var(--bg-elevated)',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border)'
            }}>
                <h3 style={{ margin: '0 0 var(--space-sm) 0', color: 'var(--text-secondary)' }}>Total Services</h3>
                <p style={{ fontSize: '2rem', fontWeight: 'bold', margin: 0, color: 'var(--primary)' }}>{totalServices}</p>
            </div>
            <div style={{
                textAlign: 'center',
                padding: 'var(--space-lg)',
                background: 'var(--bg-elevated)',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border)'
            }}>
                <h3 style={{ margin: '0 0 var(--space-sm) 0', color: 'var(--text-secondary)' }}>Operational</h3>
                <p style={{ fontSize: '2rem', fontWeight: 'bold', margin: 0, color: 'var(--success)' }}>{operationalServices}</p>
            </div>
            <div style={{
                textAlign: 'center',
                padding: 'var(--space-lg)',
                background: 'var(--bg-elevated)',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border)'
            }}>
                <h3 style={{ margin: '0 0 var(--space-sm) 0', color: 'var(--text-secondary)' }}>Errors</h3>
                <p style={{ fontSize: '2rem', fontWeight: 'bold', margin: 0, color: 'var(--error)' }}>{errorServices}</p>
            </div>
            <div style={{
                textAlign: 'center',
                padding: 'var(--space-lg)',
                background: 'var(--bg-elevated)',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border)'
            }}>
                <h3 style={{ margin: '0 0 var(--space-sm) 0', color: 'var(--text-secondary)' }}>Uptime</h3>
                <p style={{ fontSize: '2rem', fontWeight: 'bold', margin: 0, color: 'var(--primary)' }}>{uptimePercentage}%</p>
            </div>
        </div>
    );
}

function MiniServiceGrid() {
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
                if (data && Array.isArray(data)) {
                    setServices(data);
                } else {
                    setServices([]);
                }
            } catch (err) {
                setError(err.message);
            }
        };

        getServiceData();
        const intervalId = setInterval(getServiceData, 5000);
        return () => clearInterval(intervalId);
    }, []);

    if (error) {
        return <p>Error loading services: {error}</p>;
    }

    const errorServices = services.filter(s => s.error);
    const displayServices = errorServices.length > 0
        ? errorServices.slice(0, 6)
        : services.slice(0, 6);

    return (
        <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
            gap: 'var(--space-md)'
        }}>
            {displayServices.map((service, index) => (
                <div key={index} style={{
                    padding: 'var(--space-md)',
                    background: 'var(--bg-elevated)',
                    borderRadius: 'var(--radius-md)',
                    border: `1px solid ${service.error ? 'var(--error)' : 'var(--success)'}`,
                    borderLeft: `4px solid ${service.error ? 'var(--error)' : 'var(--success)'}`
                }}>
                    <h4 style={{ margin: '0 0 var(--space-sm) 0' }}>{service.name}</h4>
                    <p style={{
                        margin: 0,
                        color: service.error ? 'var(--error)' : 'var(--success)',
                        fontSize: '0.875rem'
                    }}>
                        {service.error ? '❌ Error' : '✅ Operational'}
                    </p>
                    <p style={{ margin: 'var(--space-xs) 0 0 0', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                        {service.response_time}
                    </p>
                </div>
            ))}
        </div>
    );
}

export default function OverviewView() {
    return (
        <div className="overview-view">
            <Island title="System Status">
                <QuickStats />
            </Island>
            <Island title="Service Overview">
                <MiniServiceGrid />
            </Island>
        </div>
    );
}
