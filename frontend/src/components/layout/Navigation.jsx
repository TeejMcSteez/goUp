export default function Navigation({ activeTab, onTabChange }) {
    const tabs = [
        { id: 'overview', label: 'Overview', key: '1' },
        { id: 'services', label: 'Services', key: '2' },
        { id: 'analytics', label: 'Analytics', key: '3' },
        { id: 'settings', label: 'Settings', key: '4' }
    ];

    return (
        <nav className="tab-navigation" role="tablist" aria-label="Main navigation">
            <div className="tab-list">
                {tabs.map((tab) => (
                    <button
                        key={tab.id}
                        role="tab"
                        aria-selected={activeTab === tab.id}
                        aria-controls={`${tab.id}-panel`}
                        id={`${tab.id}-tab`}
                        className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
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
