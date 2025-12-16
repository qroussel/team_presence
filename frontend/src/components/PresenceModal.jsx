import React, { useState, useEffect } from 'react';
import { format } from 'date-fns';

const PresenceModal = ({ isOpen, onClose, date, currentAM, currentPM, onSave }) => {
    const [amStatus, setAmStatus] = useState(currentAM);
    const [pmStatus, setPmStatus] = useState(currentPM);

    useEffect(() => {
        if (isOpen) {
            setAmStatus(currentAM || '');
            setPmStatus(currentPM || '');
        }
    }, [isOpen, currentAM, currentPM]);

    if (!isOpen) return null;

    const handleSave = () => {
        onSave(amStatus, pmStatus);
        onClose();
    };

    const options = [
        { label: 'Office', value: 'office', color: '#10b981' },
        { label: 'Remote', value: 'remote', color: '#3b82f6' },
        { label: 'Off', value: 'off', color: '#ef4444' },
        { label: 'None', value: '', color: '#94a3b8' },
    ];

    return (
        <div className="modal-overlay" onClick={handleSave}>
            <div className="modal-content glass-panel" onClick={(e) => e.stopPropagation()}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                    <h3 style={{ margin: 0 }}>{date ? format(date, 'MMMM d, yyyy') : 'Select Status'}</h3>
                    <button onClick={handleSave} style={{ background: 'transparent', padding: '0.5rem', color: 'var(--text-secondary)' }}>✕</button>
                </div>

                {/* Full Day Shortcuts */}
                <div style={{ marginBottom: '1.5rem', paddingBottom: '1.5rem', borderBottom: '1px solid var(--glass-border)' }}>
                    <h4 style={{ textAlign: 'center', marginBottom: '0.75rem', color: '#94a3b8', fontSize: '0.8rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Full Day (Quick Set)</h4>
                    <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                        {options.filter(o => o.value !== '').map(opt => (
                            <button
                                key={`full-${opt.value}`}
                                onClick={() => { setAmStatus(opt.value); setPmStatus(opt.value); }}
                                style={{
                                    flex: 1,
                                    background: (amStatus === opt.value && pmStatus === opt.value) ? opt.color : 'rgba(255,255,255,0.05)',
                                    border: `1px solid ${(amStatus === opt.value && pmStatus === opt.value) ? opt.color : 'rgba(255,255,255,0.1)'}`,
                                    padding: '0.5rem',
                                    color: (amStatus === opt.value && pmStatus === opt.value) ? 'white' : 'var(--text-secondary)',
                                    fontWeight: (amStatus === opt.value && pmStatus === opt.value) ? 'bold' : 'normal',
                                    fontSize: '0.85rem',
                                    borderRadius: '0.25rem'
                                }}
                            >
                                {opt.label}
                            </button>
                        ))}
                    </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
                    {/* AM Section */}
                    <div>
                        <h4 style={{ textAlign: 'center', marginBottom: '1rem', color: '#94a3b8' }}>Morning (AM)</h4>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                            {options.map(opt => (
                                <button
                                    key={`am-${opt.value}`}
                                    onClick={() => setAmStatus(opt.value)}
                                    style={{
                                        background: amStatus === opt.value ? opt.color : 'rgba(255,255,255,0.05)',
                                        border: `1px solid ${amStatus === opt.value ? opt.color : 'rgba(255,255,255,0.1)'}`,
                                        padding: '0.75rem',
                                        color: amStatus === opt.value ? 'white' : 'var(--text-secondary)',
                                        transition: 'all 0.2s',
                                        fontWeight: amStatus === opt.value ? 'bold' : 'normal'
                                    }}
                                >
                                    {opt.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    {/* PM Section */}
                    <div>
                        <h4 style={{ textAlign: 'center', marginBottom: '1rem', color: '#94a3b8' }}>Afternoon (PM)</h4>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                            {options.map(opt => (
                                <button
                                    key={`pm-${opt.value}`}
                                    onClick={() => setPmStatus(opt.value)}
                                    style={{
                                        background: pmStatus === opt.value ? opt.color : 'rgba(255,255,255,0.05)',
                                        border: `1px solid ${pmStatus === opt.value ? opt.color : 'rgba(255,255,255,0.1)'}`,
                                        padding: '0.75rem',
                                        color: pmStatus === opt.value ? 'white' : 'var(--text-secondary)',
                                        transition: 'all 0.2s',
                                        fontWeight: pmStatus === opt.value ? 'bold' : 'normal'
                                    }}
                                >
                                    {opt.label}
                                </button>
                            ))}
                        </div>
                    </div>
                </div>

                <div style={{ marginTop: '2rem', display: 'flex', justifyContent: 'flex-end' }}>
                    <button onClick={handleSave} style={{ padding: '0.75rem 2rem', fontSize: '1rem' }}>Done</button>
                </div>
            </div>
        </div>
    );
};

export default PresenceModal;
