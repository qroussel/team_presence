import React, { useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';

const ContextMenu = ({ x, y, onClose, onSelect }) => {
    console.log("ContextMenu rendering");
    const menuRef = useRef(null);

    useEffect(() => {
        const handleClickOutside = (e) => {
            if (menuRef.current && !menuRef.current.contains(e.target)) {
                onClose();
            }
        };
        document.addEventListener('click', handleClickOutside);
        return () => document.removeEventListener('click', handleClickOutside);
    }, [onClose]);

    const options = [
        { label: 'Office', value: 'office' },
        { label: 'Remote', value: 'remote' },
        { label: 'Off', value: 'off' },
        { label: 'Clear', value: '' },
    ];

    const handleSelect = (period, status) => {
        onSelect(period, status);
    };

    return createPortal(
        <div
            ref={menuRef}
            className="context-menu glass-panel"
            style={{
                position: 'fixed',
                top: y,
                left: x,
                zIndex: 1000,
                padding: '0.5rem',
                minWidth: '200px',
                background: 'var(--bg-secondary)',
                border: '1px solid var(--glass-border)',
                color: 'var(--text-primary)'
            }}
        >
            <div style={{ padding: '0.5rem', fontWeight: 'bold', borderBottom: '1px solid var(--glass-border)', marginBottom: '0.5rem' }}>
                Set Status
            </div>

            <div style={{ display: 'flex', gap: '1rem' }}>
                <div style={{ flex: 1 }}>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>AM</div>
                    {options.map(opt => (
                        <div
                            key={`am-${opt.value}`}
                            className="menu-item"
                            onClick={() => handleSelect('AM', opt.value)}
                            style={{ padding: '0.25rem', cursor: 'pointer', fontSize: '0.875rem' }}
                        >
                            {opt.label}
                        </div>
                    ))}
                </div>
                <div style={{ width: '1px', background: 'var(--glass-border)' }}></div>
                <div style={{ flex: 1 }}>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>PM</div>
                    {options.map(opt => (
                        <div
                            key={`pm-${opt.value}`}
                            className="menu-item"
                            onClick={() => handleSelect('PM', opt.value)}
                            style={{ padding: '0.25rem', cursor: 'pointer', fontSize: '0.875rem' }}
                        >
                            {opt.label}
                        </div>
                    ))}
                </div>
            </div>
        </div>,
        document.body
    );
};

export default ContextMenu;
