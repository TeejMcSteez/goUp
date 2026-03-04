import { useState } from "react";
import useErrorData from "../hooks/useErrorData.js";
import useServiceName from "../hooks/useServiceName.js";

export default function ErrorLogViewer() {
  const [limit, setLimit] = useState(100);
  const [sortOrder, setSortOrder] = useState("desc");
  const { data: errors, loading, error } = useErrorData(limit, sortOrder);
  const { formatName } = useServiceName();

  return (
    <div className="error-log-container">
      <div className="error-log-controls">
        <label htmlFor="error-limit">Show: </label>
        <select
          id="error-limit"
          value={limit}
          onChange={(e) => setLimit(Number(e.target.value))}
          className="sort-select"
          style={{ maxWidth: "120px", marginBottom: "var(--space-md)", display: "inline-block" }}
        >
          <option value={10}>10</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
          <option value={0}>All</option>
        </select>
        <label htmlFor="error-sort" style={{ marginLeft: "var(--space-md)" }}>Sort by date: </label>
        <select
          id="error-sort"
          value={sortOrder}
          onChange={(e) => setSortOrder(e.target.value)}
          className="sort-select"
          style={{ maxWidth: "120px", marginBottom: "var(--space-md)", display: "inline-block" }}
        >
          <option value="desc">Newest first</option>
          <option value="asc">Oldest first</option>
        </select>
      </div>
      <div className="error-log-viewer">
        {error ? (
          <p>Error loading error logs: {error}</p>
        ) : loading && !errors ? (
          <p>Loading errors...</p>
        ) : errors && errors.length > 0 ? (
          errors.map((errorItem, index) => (
            <div key={index} className="error-log-item">
              <div className="error-log-header">
                <strong>{formatName(errorItem.name)}</strong>
                <span>{new Date(errorItem.timestamp).toLocaleString()}</span>
              </div>
              <div className="error-log-body">
                <p>
                  <strong>Response:</strong> {errorItem.response}
                </p>
                {errorItem.data && (
                  <p>
                    <strong>Data:</strong> <pre>{errorItem.data}</pre>
                  </p>
                )}
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
