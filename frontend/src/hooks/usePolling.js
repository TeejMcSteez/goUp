import { useState, useEffect, useCallback } from 'react';

export default function usePolling(fetchFunction, interval = 5000) {
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const fetchData = useCallback(async () => {
        try {
            setLoading(true);
            const result = await fetchFunction();
            setData(result);
            setError(null);
        } catch (err) {
            setError(err.message);
            console.error("Error fetching data:", err);
        } finally {
            setLoading(false);
        }
    }, [fetchFunction]);

    useEffect(() => {
        fetchData();
        const intervalId = setInterval(fetchData, interval);
        return () => clearInterval(intervalId);
    }, [fetchData, interval]);

    return { data, loading, error, refetch: fetchData };
}
