import React, { useState, useEffect } from 'react';

const UserSettingsModal = ({ user, isOpen, onClose, onDeleteUser, teams }) => {
    if (!isOpen || !user) return null;

    const [userTeams, setUserTeams] = useState([]);
    const [loading, setLoading] = useState(false);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

    // Add member state
    const [selectedTeamToAdd, setSelectedTeamToAdd] = useState('');
    const [productivityToAdd, setProductivityToAdd] = useState(100);

    const fetchUserTeams = async () => {
        setLoading(true);
        try {
            const res = await fetch(`/api/team_members?user_id=${user.id}`);
            if (res.ok) {
                const data = await res.json();
                setUserTeams(data || []);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (isOpen && user) {
            fetchUserTeams();
            setShowDeleteConfirm(false);
        }
    }, [isOpen, user]);

    const handleUpdateProductivity = async (teamId, newProd) => {
        try {
            const res = await fetch('/api/team_members', {
                method: 'POST', // Query handles upsert
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    team_id: teamId,
                    user_id: user.id,
                    productivity: parseInt(newProd)
                })
            });
            if (res.ok) {
                fetchUserTeams();
            }
        } catch (e) {
            alert("Failed to update");
        }
    };

    const handleRemoveFromTeam = async (teamId) => {
        if (!confirm("Remove user from this team?")) return;
        try {
            const res = await fetch(`/api/team_members?team_id=${teamId}&user_id=${user.id}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                fetchUserTeams();
            }
        } catch (e) {
            alert("Failed to remove");
        }
    };

    const handleAddToTeam = async () => {
        if (!selectedTeamToAdd) return;
        try {
            const res = await fetch('/api/team_members', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    team_id: parseInt(selectedTeamToAdd),
                    user_id: user.id,
                    productivity: parseInt(productivityToAdd)
                })
            });
            if (res.ok) {
                fetchUserTeams();
                setSelectedTeamToAdd('');
                setProductivityToAdd(100);
            }
        } catch (e) {
            alert("Failed to add");
        }
    };

    const handleDeleteUser = () => {
        onDeleteUser(user.id);
        onClose();
    };

    return (
        <div style={{
            position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000
        }}>
            <div className="glass-panel" style={{ width: '500px', maxHeight: '90vh', overflowY: 'auto', position: 'relative' }}>
                <button onClick={onClose} style={{ position: 'absolute', top: '10px', right: '10px', background: 'transparent', border: 'none', color: 'white', cursor: 'pointer' }}>X</button>

                <h2 style={{ marginTop: 0 }}>Settings: {user.name}</h2>

                {/* Team Memberships */}
                <div style={{ marginBottom: '2rem' }}>
                    <h3>Team Memberships</h3>
                    {loading ? <div>Loading...</div> : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                            {userTeams.length === 0 && <span style={{ color: 'var(--text-secondary)' }}>Not in any teams.</span>}
                            {userTeams.map(tm => (
                                <div key={tm.team_id} style={{ display: 'flex', alignItems: 'center', gap: '1rem', background: 'rgba(255,255,255,0.05)', padding: '0.5rem', borderRadius: '4px' }}>
                                    <span style={{ flex: 1, fontWeight: 'bold' }}>{tm.team_name || `Team ${tm.team_id}`}</span>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <label>Prod %:</label>
                                        <input
                                            type="number"
                                            defaultValue={tm.productivity}
                                            onBlur={(e) => handleUpdateProductivity(tm.team_id, e.target.value)}
                                            style={{ width: '60px', padding: '0.25rem', background: 'rgba(0,0,0,0.3)', border: '1px solid var(--glass-border)', color: 'white', borderRadius: '4px' }}
                                        />
                                    </div>
                                    <button onClick={() => handleRemoveFromTeam(tm.team_id)} style={{ background: '#ef4444', border: 'none', color: 'white', padding: '0.25rem 0.5rem', borderRadius: '4px', cursor: 'pointer' }}>Remove</button>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* Add to Team */}
                <div style={{ marginBottom: '2rem', padding: '1rem', background: 'rgba(255,255,255,0.03)', borderRadius: '8px' }}>
                    <h4>Add to Team</h4>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <select
                            value={selectedTeamToAdd}
                            onChange={e => setSelectedTeamToAdd(e.target.value)}
                            style={{ flex: 1, background: 'rgba(0,0,0,0.3)', border: '1px solid var(--glass-border)', color: 'white', padding: '0.5rem', borderRadius: '4px' }}
                        >
                            <option value="">Select Team...</option>
                            {teams.filter(t => !userTeams.find(ut => ut.team_id === t.id)).map(t => (
                                <option key={t.id} value={t.id}>{t.name}</option>
                            ))}
                        </select>
                        <input
                            type="number"
                            value={productivityToAdd}
                            onChange={e => setProductivityToAdd(e.target.value)}
                            placeholder="%"
                            style={{ width: '60px', padding: '0.5rem', background: 'rgba(0,0,0,0.3)', border: '1px solid var(--glass-border)', color: 'white', borderRadius: '4px' }}
                        />
                        <button onClick={handleAddToTeam} disabled={!selectedTeamToAdd} style={{ cursor: 'pointer' }}>Add</button>
                    </div>
                </div>

                {/* Danger Zone */}
                <div style={{ borderTop: '1px solid #ef4444', paddingTop: '1rem' }}>
                    {!showDeleteConfirm ? (
                        <button
                            onClick={() => setShowDeleteConfirm(true)}
                            style={{ width: '100%', background: 'rgba(239, 68, 68, 0.2)', color: '#ef4444', border: '1px solid #ef4444', padding: '0.75rem', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold' }}
                        >
                            Delete User
                        </button>
                    ) : (
                        <div style={{ background: 'rgba(239, 68, 68, 0.1)', padding: '1rem', borderRadius: '6px', textAlign: 'center' }}>
                            <p style={{ color: '#ef4444', fontWeight: 'bold', marginBottom: '1rem' }}>Are you sure? This action is irreversible.</p>
                            <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
                                <button
                                    onClick={() => setShowDeleteConfirm(false)}
                                    style={{ background: 'transparent', border: '1px solid var(--glass-border)', color: 'white', padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer' }}
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleDeleteUser}
                                    style={{ background: '#ef4444', border: 'none', color: 'white', padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}
                                >
                                    Confirm Delete
                                </button>
                            </div>
                        </div>
                    )}
                </div>

            </div>
        </div>
    );
};

export default UserSettingsModal;
