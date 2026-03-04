import useServiceName from "../hooks/useServiceName.js";

export default function DisplaySettings() {
  const { prettify, toggle } = useServiceName();

  return (
    <div>
      <label
        style={{
          display: "flex",
          alignItems: "center",
          gap: "var(--space-md)",
          cursor: "pointer",
        }}
      >
        <input type="checkbox" checked={prettify} onChange={toggle} />
        Prettify service names (strip URL prefixes, replace underscores with spaces)
      </label>
    </div>
  );
}
