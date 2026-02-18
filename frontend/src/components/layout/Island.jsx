export default function Island({ title, children, headerAction }) {
  return (
    <div className="island">
      {title && (
        <div className="island-header">
          <h2 className="island-title">{title}</h2>
          {headerAction && <div className="island-action">{headerAction}</div>}
        </div>
      )}
      <div className="island-content">{children}</div>
    </div>
  );
}
