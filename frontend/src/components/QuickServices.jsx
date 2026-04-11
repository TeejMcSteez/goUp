import { useState, useEffect } from "react";

export default function QuickServices() {
  const [qs, setQs] = useState([]);
  const [error, setError] = useState(null);

  useEffect(() => {
    const getDownServices = async () => {
      try {
        const res = await fetch("/api/status");
        if (!res.ok) {
          throw new Error(`Server error: ${res.status}`);
        }
        const data = await res.json();
        if (Array.isArray(data)) {
          setQs(data);
        } else if (data === null) {
          // Go server returns null for no downed services
          // This sets downed services to 0 (empty array) when there are no downed services
          // So that state change will be properly hydrated throughout QuickServices view
          setQs([]);
        } else {
          console.error(`Expected array from /api/status, got: ${data}`);
        }
        setError(null);
      } catch (err) {
        setError(err.message);
        console.error("Error fetching quick service data:", err);
      }
    };

    getDownServices();

    const intervalId = setInterval(getDownServices, 5000);

    return () => clearInterval(intervalId);
  }, []);

  let content;
  if (qs.length === 0) {
    content = (
      <a
        href="#cards"
        className="no-underline text-center text-lg text-fg px-6 py-2 rounded-lg transition-colors duration-200 hover:bg-hover"
      >
        All Systems Operational ✅
      </a>
    );
  } else {
    content = (
      <a
        href="#cards"
        className="no-underline text-center text-lg text-fg px-6 py-2 rounded-lg transition-colors duration-200 hover:bg-hover"
      >
        {qs.length} Errors Detected ❌
      </a>
    );
  }

  return (
    <div className="flex flex-row items-center justify-center text-center p-4 border-b border-border bg-surface">
      {error ? (
        <span className="text-error text-xs">
          Status unavailable: {error}
        </span>
      ) : (
        content
      )}
    </div>
  );
}
