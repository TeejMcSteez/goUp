export default function StatusMessage({ message, isError }) {
  if (!message) return null;
  return (
    <p
      className={`text-[0.9rem] m-0 px-4 py-2 rounded-lg ${
        isError
          ? "text-error bg-error/10 border border-error/30"
          : "text-success bg-success/10 border border-success/30"
      }`}
    >
      {message}
    </p>
  );
}
