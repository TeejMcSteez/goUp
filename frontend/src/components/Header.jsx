import QuickServices from "./QuickServices.jsx";

export default function Header() {
    return (
        <header style={{
            position: 'sticky',
            top: 0,
            zIndex: 200,
            background: 'var(--bg-primary)',
            borderBottom: '1px solid var(--border)'
        }}>
            <div style={{
                textAlign: 'center',
                padding: 'var(--space-xl) var(--space-md) var(--space-md)',
                background: 'var(--bg-primary)'
            }}>
                <h1 style={{ margin: 0, fontSize: '2.5rem', color: 'var(--primary)' }}>GoUp</h1>
                <h2 style={{ margin: 'var(--space-sm) 0 0', fontSize: '1rem', color: 'var(--text-secondary)', fontWeight: 'normal' }}>
                    Service Monitoring Dashboard
                </h2>
            </div>
            <QuickServices />
        </header>
    );
}