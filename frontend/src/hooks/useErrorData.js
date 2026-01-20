import { useCallback } from 'react';
import usePolling from './usePolling.js';

export default function useErrorData() {
    const fetchErrors = useCallback(async () => {
        const res = await fetch("/api/errors");
        if (!res.ok) {
            throw new Error(`HTTP error! status: ${res.status}`);
        }
        const data = await res.json();
        return data || [];
    }, []);

    return usePolling(fetchErrors, 5000);
}
