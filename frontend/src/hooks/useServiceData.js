import { useCallback } from 'react';
import usePolling from './usePolling.js';

export default function useServiceData() {
    const fetchServices = useCallback(async () => {
        const res = await fetch("/api");
        if (!res.ok) {
            throw new Error(`Server error: ${res.status}`);
        }
        const data = await res.json();
        if (data && Array.isArray(data)) {
            return data;
        } else {
            console.error("Expected array from /api, got:", data);
            return [];
        }
    }, []);

    return usePolling(fetchServices, 5000);
}
