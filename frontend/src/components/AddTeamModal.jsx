import React, { useState } from 'react';

export default function AddTeamModal({ isOpen, onClose, users, onTeamCreated }) {
    const [teamName, setTeamName] = useState('');
    const [selectedMembers, setSelectedMembers] = useState({}); // { userId: { selected: bool, productivity: int } }
    const [isSubmitting, setIsSubmitting] = useState(false);

    if (!isOpen) return null;

    const handleMemberToggle = (userId) => {
        setSelectedMembers(prev => {
            const current = prev[userId] || { selected: false, productivity: 100 };
            return {
                ...prev,
                [userId]: { ...current, selected: !current.selected }
            };
        });
    };

    const handleProductivityChange = (userId, value) => {
        const val = Math.max(0, Math.min(100, parseInt(value) || 0));
        setSelectedMembers(prev => ({
            ...prev,
            [userId]: { ...prev[userId], productivity: val }
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!teamName.trim()) return;

        setIsSubmitting(true);
        try {
            // 1. Create Team
            const teamRes = await fetch('/api/teams', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: teamName })
            });

            if (!teamRes.ok) throw new Error('Failed to create team');
            const team = await teamRes.json();

            // 2. Add Selected Members
            const memberPromises = Object.entries(selectedMembers)
                .filter(([_, data]) => data.selected)
                .map(([userId, data]) => {
                    return fetch('/api/team_members', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            team_id: team.id,
                            user_id: parseInt(userId),
                            productivity: data.productivity
                        })
                    });
                });

            await Promise.all(memberPromises);

            onTeamCreated();
            onClose();
            // Reset form
            setTeamName('');
            setSelectedMembers({});
        } catch (error) {
            console.error(error);
            alert('Error creating team: ' + error.message);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div style={{ padding: '1.5rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                        <h2 style={{ margin: 0 }}>Add New Team</h2>
                        <button onClick={onClose} style={{ background: 'transparent', color: 'var(--text-secondary)', padding: '0.25rem' }}>✕</button>
                    </div>

                    <form onSubmit={handleSubmit}>
                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: 'var(--text-secondary)' }}>Team Name</label>
                            <input
                                required
                                value={teamName}
                                onChange={e => setTeamName(e.target.value)}
                                placeholder="e.g. Engineering Squad A"
                                style={{ width: '100%', boxSizing: 'border-box' }}
                            />
                        </div>

                        <div style={{ marginBottom: '1.5rem' }}>
                            <label style={{ display: 'block', marginBottom: '0.5rem', color: 'var(--text-secondary)' }}>Select Members</label>
                            <div style={{
                                maxHeight: '300px',
                                overflowY: 'auto',
                                border: '1px solid var(--glass-border)',
                                borderRadius: '0.5rem',
                                background: 'rgba(0,0,0,0.2)'
                            }}>
                                {users.map(user => {
                                    const memberData = selectedMembers[user.id] || { selected: false, productivity: 100 };
                                    return (
                                        <div key={user.id} style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            padding: '0.75rem',
                                            borderBottom: '1px solid var(--glass-border)',
                                            gap: '1rem'
                                        }}>
                                            <input
                                                type="checkbox"
                                                checked={memberData.selected}
                                                onChange={() => handleMemberToggle(user.id)}
                                                style={{ width: 'auto' }}
                                            />
                                            <div style={{ flex: 1 }}>
                                                <div style={{ fontWeight: 500 }}>{user.name}</div>
                                                <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{user.email}</div>
                                            </div>
                                            {memberData.selected && (
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                    <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>Prod %</span>
                                                    <input
                                                        type="number"
                                                        min="0"
                                                        max="100"
                                                        value={memberData.productivity}
                                                        onChange={(e) => handleProductivityChange(user.id, e.target.value)}
                                                        style={{ width: '60px', padding: '0.25rem' }}
                                                    />
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>

                        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem' }}>
                            <button type="button" onClick={onClose} style={{ background: 'transparent', border: '1px solid var(--glass-border)' }}>Cancel</button>
                            <button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? 'Creating...' : 'Create Team'}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
}
