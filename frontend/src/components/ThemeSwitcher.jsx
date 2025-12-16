import React from 'react';
import { useTheme } from '../context/ThemeContext';

export default function ThemeSwitcher() {
    const { theme, setTheme } = useTheme();

    return (
        <div style={{
            display: 'flex',
            background: 'rgba(255,255,255,0.1)',
            padding: '4px',
            borderRadius: '8px',
            gap: '4px'
        }}>
            {[
                { id: 'light', icon: '☀️', label: 'Light' },
                { id: 'system', icon: '🖥️', label: 'System' },
                { id: 'dark', icon: '🌙', label: 'Dark' },
            ].map((mode) => (
                <button
                    key={mode.id}
                    onClick={() => setTheme(mode.id)}
                    title={mode.label}
                    style={{
                        background: theme === mode.id ? 'var(--accent)' : 'transparent',
                        color: theme === mode.id ? 'white' : 'var(--text-secondary)',
                        border: 'none',
                        padding: '4px 8px',
                        borderRadius: '6px',
                        cursor: 'pointer',
                        fontSize: '1.2rem',
                        lineHeight: 1,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        transition: 'all 0.2s',
                        minWidth: '36px'
                    }}
                >
                    {mode.icon}
                </button>
            ))}
        </div>
    );
}
