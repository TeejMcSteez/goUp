import { useState } from "react";
import useErrorData from "../hooks/useErrorData";
import useServiceName from "../hooks/useServiceName";

export default function ErrorLogViewer() {
  const [limit, setLimit] = useState(100);
  const [sortOrder, setSortOrder] = useState("desc");
  const { data: errors, loading, error } = useErrorData(limit, sortOrder);
  const { formatName } = useServiceName();

  const selectClass =
    "max-w-[120px] px-4 py-3 mb-4 inline-block rounded-lg border border-border bg-surface text-fg text-base transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.3)]";

  return (
    <div className="w-full">
      <div className="flex items-center gap-2 mb-4 flex-wrap">
        <label htmlFor="error-limit">Show: </label>
        <select
          id="error-limit"
          value={limit}
          onChange={(e) => setLimit(Number(e.target.value))}
          className={selectClass}
        >
          <option value={10}>10</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
          <option value={0}>All</option>
        </select>
        <label htmlFor="error-sort" className="ml-4">Sort by date: </label>
        <select
          id="error-sort"
          value={sortOrder}
          onChange={(e) => setSortOrder(e.target.value)}
          className={selectClass}
        >
          <option value="desc">Newest first</option>
          <option value="asc">Oldest first</option>
        </select>
      </div>
      <div className="max-h-[400px] overflow-y-auto pr-4 scrollbar-custom">
        {error ? (
          <p>Error loading error logs: {error}</p>
        ) : loading && !errors ? (
          <p>Loading errors...</p>
        ) : errors && errors.length > 0 ? (
          errors.map((errorItem) => (
            <div key={`${errorItem.name}-${errorItem.timestamp}`} className="p-4 border-b border-border last:border-b-0">
              <div className="flex justify-between items-center mb-2 text-[1.1rem]">
                <strong>{formatName(errorItem.name)}</strong>
                <span className="text-[0.9rem] text-muted">
                  {new Date(errorItem.timestamp).toLocaleString()}
                </span>
              </div>
              <div className="text-[0.95rem] text-fg">
                <p className="my-2">
                  <strong>Response:</strong> {errorItem.response}
                </p>
                {errorItem.data && (
                  <p className="my-2">
                    <strong>Data:</strong>{" "}
                    <pre className="bg-hover p-2 rounded whitespace-pre-wrap break-words">
                      {errorItem.data}
                    </pre>
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
