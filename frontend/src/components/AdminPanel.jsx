import React, { useState, useEffect } from 'react'

import PropTypes from 'prop-types'

export default function AdminPanel({ onBack }) {
    const [teams, setTeams] = useState([])
    const [users, setUsers] = useState([])
    const [selectedTeamId, setSelectedTeamId] = useState(null)


    // Team Member Management State
    const [teamMembers, setTeamMembers] = useState([])
    const [selectedUserToAdd, setSelectedUserToAdd] = useState('')
    const [productivityToAdd, setProductivityToAdd] = useState(100)

    // Create Team State
    const [newTeamName, setNewTeamName] = useState('')

    useEffect(() => {
        fetchTeams()
        fetchUsers()
    }, [])

    useEffect(() => {
        if (selectedTeamId) {
            fetchTeamMembers(selectedTeamId)
        } else {
            setTeamMembers([])
        }
    }, [selectedTeamId])

    const fetchTeams = async () => {
        try {
            const res = await fetch('/api/teams')
            if (res.ok) setTeams(await res.json() || [])
        } catch (e) {
            console.error(e)
        }
    }

    const fetchUsers = async () => {
        try {
            const res = await fetch('/api/users')
            if (res.ok) setUsers(await res.json() || [])
        } catch (e) {
            console.error(e)
        }
    }

    const fetchTeamMembers = async (teamId) => {
        try {
            const res = await fetch(`/api/team_members?team_id=${teamId}`)
            if (res.ok) setTeamMembers(await res.json() || [])
        } catch (e) {
            console.error(e)
        }
    }

    const handleCreateTeam = async (e) => {
        e.preventDefault()
        if (!newTeamName) return
        try {
            const res = await fetch('/api/teams', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: newTeamName })
            })
            if (res.ok) {
                setNewTeamName('')
                fetchTeams()
            }
        } catch (e) { console.error(e) }
    }

    const handleDeleteTeam = async (teamId) => {
        if (!globalThis.confirm('Are you sure you want to delete this team? All members will be removed from it.')) return
        try {
            const res = await fetch(`/api/teams?id=${teamId}`, { method: 'DELETE' })
            if (res.ok) {
                fetchTeams()
                if (selectedTeamId === teamId) setSelectedTeamId(null)
            }
        } catch (e) { console.error(e) }
    }

    const handleAddMember = async (e) => {
        e.preventDefault()
        if (!selectedTeamId || !selectedUserToAdd) return

        try {
            const res = await fetch('/api/team_members', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    team_id: selectedTeamId,
                    user_id: Number.parseInt(selectedUserToAdd, 10),
                    productivity: Number.parseInt(productivityToAdd, 10)
                })
            })
            if (res.ok) {
                fetchTeamMembers(selectedTeamId)
                // Reset form? maybe keep productivity
            }
        } catch (e) { console.error(e) }
    }

    const handleRemoveMember = async (userId) => {
        if (!selectedTeamId) return
        try {
            const res = await fetch(`/api/team_members?team_id=${selectedTeamId}&user_id=${userId}`, { method: 'DELETE' })
            if (res.ok) fetchTeamMembers(selectedTeamId)
        } catch (e) { console.error(e) }
    }

    const handleUpdateProductivity = async (userId, newProd) => {
        try {
            const res = await fetch('/api/team_members', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    team_id: selectedTeamId,
                    user_id: userId,
                    productivity: Number.parseInt(newProd, 10)
                })
            })
            if (res.ok) {
                // Optimistic update or refetch
                fetchTeamMembers(selectedTeamId)
            }
        } catch (e) { console.error(e) }
    }

    return (
        <div style={{ padding: '2rem', maxWidth: '1200px', margin: '0 auto' }}>
            <button
                onClick={onBack}
                style={{ marginBottom: '1rem', background: 'transparent', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem' }}
            >
                ← Back to App
            </button>

            <h2 style={{ marginBottom: '2rem' }}>Team Management</h2>

            <div style={{ display: 'grid', gridTemplateColumns: 'minmax(300px, 1fr) 2fr', gap: '2rem' }}>

                {/* Teams List */}
                <div className="glass-panel">
                    <h3 style={{ marginTop: 0 }}>Teams</h3>

                    <form onSubmit={handleCreateTeam} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
                        <input
                            value={newTeamName}
                            onChange={e => setNewTeamName(e.target.value)}
                            placeholder="New Team Name"
                            style={{ flex: 1 }}
                        />
                        <button type="submit">Add</button>
                    </form>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                        {teams.map(team => (
                            <div
                                key={team.id}
                                style={{
                                    display: 'flex',
                                    gap: '0.5rem',
                                    borderRadius: '0.5rem',
                                    background: selectedTeamId === team.id ? 'var(--selected-item-bg)' : 'rgba(255,255,255,0.05)',
                                    padding: '0.5rem',
                                    alignItems: 'center'
                                }}
                            >
                                <button
                                    onClick={() => setSelectedTeamId(team.id)}
                                    style={{
                                        flex: 1,
                                        background: 'transparent',
                                        border: 'none',
                                        color: 'inherit',
                                        textAlign: 'left',
                                        cursor: 'pointer',
                                        padding: '0.5rem',
                                        fontSize: '1rem',
                                        display: 'flex',
                                        alignItems: 'center',
                                        height: '100%'
                                    }}
                                >
                                    {team.name}
                                </button>
                                <button
                                    onClick={(e) => { e.stopPropagation(); handleDeleteTeam(team.id) }}
                                    aria-label={`Delete ${team.name}`}
                                    style={{
                                        background: 'var(--danger-bg, #ef4444)',
                                        color: 'white',
                                        border: 'none',
                                        padding: '0.25rem 0.5rem',
                                        borderRadius: '4px',
                                        fontSize: '0.8rem',
                                        cursor: 'pointer',
                                        flexShrink: 0
                                    }}
                                >
                                    Delete
                                </button>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Members Management */}
                <div className="glass-panel">
                    {selectedTeamId ? (
                        <>
                            <h3 style={{ marginTop: 0 }}>
                                Members of {teams.find(t => t.id === selectedTeamId)?.name}
                            </h3>

                            <div style={{ marginBottom: '1.5rem', padding: '1rem', background: 'rgba(255,255,255,0.05)', borderRadius: '0.5rem' }}>
                                <h4 style={{ marginTop: 0 }}>Add Member</h4>
                                <form onSubmit={handleAddMember} style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
                                    <select
                                        value={selectedUserToAdd}
                                        onChange={e => setSelectedUserToAdd(e.target.value)}
                                        style={{ padding: '0.5rem', borderRadius: '4px', background: 'var(--bg-secondary)', color: 'var(--text-primary)', border: '1px solid var(--glass-border)' }}
                                    >
                                        <option value="">Select User</option>
                                        {users.map(u => (
                                            <option key={u.id} value={u.id}>{u.name} ({u.email})</option>
                                        ))}
                                    </select>
                                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                        <span>Productivity:</span>
                                        <input
                                            type="number"
                                            min="0" max="100"
                                            value={productivityToAdd}
                                            onChange={e => setProductivityToAdd(e.target.value)}
                                            style={{ width: '60px' }}
                                        />
                                        <span>%</span>
                                    </label>
                                    <button type="submit" disabled={!selectedUserToAdd}>Add to Team</button>
                                </form>
                            </div>

                            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                {teamMembers.length === 0 && <span style={{ color: 'var(--text-secondary)' }}>No members in this team.</span>}
                                {teamMembers.map(tm => (
                                    <div key={tm.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.75rem', background: 'rgba(255,255,255,0.05)', borderRadius: '0.5rem' }}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                            <div style={{ width: '32px', height: '32px', borderRadius: '50%', background: 'linear-gradient(135deg, #6366f1, #ec4899)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.8rem', fontWeight: 'bold' }}>
                                                {tm.user_name ? tm.user_name.charAt(0).toUpperCase() : '?'}
                                            </div>
                                            <div>
                                                <div>{tm.user_name || `User ${tm.user_id}`}</div>
                                            </div>
                                        </div>

                                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                                            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>Prod:</span>
                                                <input
                                                    type="number"
                                                    value={tm.productivity}
                                                    onChange={e => handleUpdateProductivity(tm.user_id, e.target.value)}
                                                    style={{ width: '50px', padding: '2px' }}
                                                />
                                                <span style={{ fontSize: '0.8rem' }}>%</span>
                                            </label>
                                            <button
                                                onClick={() => handleRemoveMember(tm.user_id)}
                                                style={{ background: 'transparent', border: '1px solid var(--danger-bg, #ef4444)', color: 'var(--danger-bg, #ef4444)', padding: '0.25rem 0.5rem', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem' }}
                                            >
                                                Remove
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </>
                    ) : (
                        <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-secondary)' }}>
                            Select a team to manage members
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}

AdminPanel.propTypes = {
    onBack: PropTypes.func.isRequired
}

