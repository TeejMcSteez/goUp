export default function StatusMessage({ message, isError }) {
  if (!message) return null;
  return (
    <p className={`config-status ${isError ? "config-status--error" : "config-status--success"}`}>
      {message}
    </p>
  );
}
