import { useState } from "react";

export default function ServiceCard({ service }) {
  const { name, response, response_time, data, error } = service;
  const [showApiResponse, setShowApiResponse] = useState(false);
  const [showFullHttpResponse, setShowFullHttpResponse] = useState(false);

  const toggleApiResponse = () => {
    setShowApiResponse(!showApiResponse);
  };

  const toggleFullHttpResponse = () => {
    setShowFullHttpResponse(!showFullHttpResponse);
  };

  const isLongHttpResponse = error && response && response.length > 3;

  return (
    <div className="card">
      <h3 className="svcName">Name: {name}</h3>
      <p>Status: {error ? "❌ Error" : "✅ Operational"}</p>
      <p>Response Time: {response_time}</p>
      {isLongHttpResponse ? (
        <>
          <h2 onClick={toggleFullHttpResponse} style={{ cursor: "pointer" }}>
            HTTP Error Response {showFullHttpResponse ? "▲" : "▼"}
          </h2>
          {showFullHttpResponse && <div className="svcHttpRes">{response}</div>}
        </>
      ) : (
        <p className="svcHttpRes">HTTP Response: {response}</p>
      )}
      <h2 onClick={toggleApiResponse} style={{ cursor: "pointer" }}>
        API Response {showApiResponse ? "▲" : "▼"}
      </h2>
      {showApiResponse && (
        <div className="svcData">
          {data ? data : "No API setup in configuration"}
        </div>
      )}
    </div>
  );
}
