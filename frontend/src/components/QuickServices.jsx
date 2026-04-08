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
      <a href="#cards" id="quickServicesCard">
        All Systems Operational ✅
      </a>
    );
  } else {
    content = (
      <a href="#cards" id="quickServicesCard">
        {qs.length} Errors Detected ❌
      </a>
    );
  }

  return (
    <div id="quickServices">
      {error ? (
        <span style={{ color: "var(--error)", fontSize: "0.75rem" }}>
          Status unavailable: {error}
        </span>
      ) : (
        content
      )}
    </div>
  );
}
