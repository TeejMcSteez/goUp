import useErrorData from "../hooks/useErrorData.js";

export default function ErrorLogViewer() {
  const { data: errors, loading, error } = useErrorData();

  return (
    <div className="error-log-container">
      <div className="error-log-viewer">
        {error ? (
          <p>Error loading error logs: {error}</p>
        ) : loading && !errors ? (
          <p>Loading errors...</p>
        ) : errors && errors.length > 0 ? (
          errors.map((errorItem, index) => (
            <div key={index} className="error-log-item">
              <div className="error-log-header">
                <strong>{errorItem.name}</strong>
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
