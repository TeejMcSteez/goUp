export default function ServiceControls({
  searchTerm,
  setSearchTerm,
  sortKey,
  setSortKey,
}) {
  const inputClass =
    "w-full max-w-[600px] px-4 py-3 mb-6 rounded-lg border border-border bg-surface text-fg text-base transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.3)] placeholder:text-muted sm:max-w-full";

  return (
    <div className="flex flex-col sm:flex-row items-center justify-center gap-4 flex-wrap">
      <input
        type="text"
        placeholder="Search services..."
        className={inputClass}
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        aria-label="Search services"
      />
      <select
        value={sortKey}
        onChange={(e) => setSortKey(e.target.value)}
        className={inputClass}
        aria-label="Sort services"
      >
        <option value="name">Sort by Name</option>
        <option value="status">Sort by Status</option>
      </select>
    </div>
  );
}
