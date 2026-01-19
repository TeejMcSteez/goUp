import { useState, useEffect } from "react";

export default function ErrorLogViewer() {
    const [errors, setErrors] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        async function fetchErrors() {
            try {
                setLoading(true);
                const response = await fetch("/api/errors");
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                setErrors(data || []);
            } catch (error) {
                console.error("Could not fetch errors:", error);
                setErrors([]);
            } finally {
                setLoading(false);
            }
        }

        fetchErrors();
    }, []);

    return (
        <div className="error-log-container">
            <h2>Error Log</h2>
            <div className="error-log-viewer">
                {loading ? (
                    <p>Loading errors...</p>
                ) : errors.length > 0 ? (
                    errors.map((error, index) => (
                        <div key={index} className="error-log-item">
                            <div className="error-log-header">
                                <strong>{error.name}</strong>
                                <span>{new Date(error.timestamp).toLocaleString()}</span>
                            </div>
                            <div className="error-log-body">
                                <p><strong>Response:</strong> {error.response}</p>
                                {error.data && <p><strong>Data:</strong> <pre>{error.data}</pre></p>}
                            </div>
                        </div>
                    ))
                ) : (
                    <p>No errors to display.</p>
                )}
            </div>
        </div>
    );
}
