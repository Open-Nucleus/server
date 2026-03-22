import { useNavigate } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

interface Breadcrumb {
  label: string;
  path?: string; // if undefined, it's the current page (not clickable)
}

interface PageHeaderProps {
  title: string;
  breadcrumbs?: Breadcrumb[];
  actions?: React.ReactNode;
}

export function PageHeader({ title, breadcrumbs, actions }: PageHeaderProps) {
  const navigate = useNavigate();

  return (
    <div style={{ marginBottom: '20px' }}>
      {/* Breadcrumbs */}
      {breadcrumbs && breadcrumbs.length > 0 && (
        <nav style={{ display: 'flex', alignItems: 'center', gap: '4px', marginBottom: '8px' }}>
          {breadcrumbs.map((crumb, i) => (
            <span key={i} style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              {i > 0 && <ChevronRight size={12} style={{ color: 'var(--color-muted)' }} />}
              {crumb.path ? (
                <button
                  type="button"
                  onClick={() => navigate({ to: crumb.path! })}
                  style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    fontFamily: 'var(--font-mono)',
                    fontSize: '11px',
                    textTransform: 'uppercase',
                    letterSpacing: '1px',
                    color: 'var(--color-muted)',
                    textDecoration: 'none',
                    padding: 0,
                  }}
                  onMouseOver={(e) => (e.currentTarget.style.color = 'var(--color-ink)')}
                  onMouseOut={(e) => (e.currentTarget.style.color = 'var(--color-muted)')}
                >
                  {crumb.label}
                </button>
              ) : (
                <span style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  textTransform: 'uppercase',
                  letterSpacing: '1px',
                  color: 'var(--color-ink)',
                }}>
                  {crumb.label}
                </span>
              )}
            </span>
          ))}
        </nav>
      )}

      {/* Title row */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h1 style={{
          fontSize: '20px',
          fontWeight: 700,
          margin: 0,
          fontFamily: 'var(--font-mono)',
          letterSpacing: '-0.3px',
          color: 'var(--color-ink)',
        }}>
          {title}
        </h1>
        {actions && <div style={{ display: 'flex', gap: '8px' }}>{actions}</div>}
      </div>
    </div>
  );
}
