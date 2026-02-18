export default function ServiceControls({
  searchTerm,
  setSearchTerm,
  sortKey,
  setSortKey,
}) {
  return (
    <div className="controls">
      <input
        type="text"
        placeholder="Search services..."
        className="search-bar"
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        aria-label="Search services"
      />
      <select
        value={sortKey}
        onChange={(e) => setSortKey(e.target.value)}
        className="sort-select"
        aria-label="Sort services"
      >
        <option value="name">Sort by Name</option>
        <option value="status">Sort by Status</option>
      </select>
    </div>
  );
}
