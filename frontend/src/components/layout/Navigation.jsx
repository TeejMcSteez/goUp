export default function Navigation({ activeTab, onTabChange }) {
  const tabs = [
    { id: "overview", label: "Overview", key: "1" },
    { id: "services", label: "Services", key: "2" },
    { id: "analytics", label: "Analytics", key: "3" },
    { id: "settings", label: "Settings", key: "4" },
  ];

  return (
    <nav
      className="sticky top-0 z-[100] bg-app-bg border-b border-border px-4"
      role="tablist"
      aria-label="Main navigation"
    >
      <div className="flex gap-1 max-w-[1400px] mx-auto overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={activeTab === tab.id}
            aria-controls={`${tab.id}-panel`}
            id={`${tab.id}-tab`}
            className={`relative px-8 py-6 bg-transparent border-none text-base font-medium cursor-pointer whitespace-nowrap transition-colors duration-200 hover:text-fg focus-visible:outline-2 focus-visible:outline focus-visible:outline-primary focus-visible:-outline-offset-2 sm:px-6 sm:py-4 sm:text-sm ${
              activeTab === tab.id ? "text-primary tab-active" : "text-muted"
            }`}
            onClick={() => onTabChange(tab.id)}
            tabIndex={activeTab === tab.id ? 0 : -1}
          >
            {tab.label}
          </button>
        ))}
      </div>
    </nav>
  );
}
