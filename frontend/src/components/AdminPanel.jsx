import React, { useState, useEffect, useMemo } from 'react'
import { apiFetch } from '../utils/api';
import PropTypes from 'prop-types'
import AddMemberModal from './AddMemberModal'
import ConfirmationModal from './ConfirmationModal'

export default function AdminPanel({ onBack, currentUser }) {
    const [teams, setTeams] = useState([])
    const [users, setUsers] = useState([])
    const [selectedTeamId, setSelectedTeamId] = useState(null)


    // Team Member Management State
    const [teamMembers, setTeamMembers] = useState([])
    const [isAddMemberModalOpen, setIsAddMemberModalOpen] = useState(false)

    // Confirmation Modal State
    const [confirmModal, setConfirmModal] = useState({
        isOpen: false,
        title: '',
        message: '',
        onConfirm: () => { },
        confirmText: 'Confirm',
        confirmColor: 'var(--accent)'
    })

    // Create Team State
    const [newTeamName, setNewTeamName] = useState('')

    // Filter teams based on the role returned by the API
    const visibleTeams = useMemo(() => {
        if (!currentUser) return []
        if (currentUser.role === 'Admin' || currentUser.role === 'admin') return teams
        return teams.filter(t => t.role === 'Owner')
    }, [teams, currentUser])



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
            const res = await apiFetch('/api/teams');
            if (res.ok) {
                const data = await res.json();
                setTeams(data || []);
            }
        } catch (e) {
            console.error(e)
        }
    }

    const fetchUsers = async () => {
        try {
            const res = await apiFetch('/api/users')
            if (res.ok) setUsers(await res.json() || [])
        } catch (e) {
            console.error(e)
        }
    }

    const fetchTeamMembers = async (teamId) => {
        try {
            const res = await apiFetch(`/api/team_members?team_id=${teamId}`)
            if (res.ok) setTeamMembers(await res.json() || [])
        } catch (e) {
            console.error(e)
        }
    }

    const handleCreateTeam = async (e) => {
        e.preventDefault()
        if (!newTeamName) return
        try {
            const res = await apiFetch('/api/teams', {
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

    const handleDeleteClick = (team) => {
        setConfirmModal({
            isOpen: true,
            title: 'Delete Team',
            message: `Are you sure you want to delete ${team.name}? This action cannot be undone.`,
            confirmText: 'Delete',
            confirmColor: '#ef4444',
            onConfirm: () => confirmDeleteTeam(team.id)
        })
    }

    const confirmDeleteTeam = async (teamId) => {
        try {
            const res = await apiFetch(`/api/teams?id=${teamId}`, { method: 'DELETE' })
            if (res.ok) {
                setSelectedTeamId(null)
                fetchTeams()
            }
        } catch (e) { console.error(e) }
    }

    const handleAddMember = async (userId, productivity) => {
        if (!selectedTeamId) return

        try {
            const res = await apiFetch('/api/team_members', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    team_id: selectedTeamId,
                    user_id: userId,
                    productivity: productivity
                })
            })
            if (res.ok) {
                fetchTeamMembers(selectedTeamId)
            }
        } catch (e) { console.error(e) }
    }

    const handleRemoveMemberClick = (userId) => {
        setConfirmModal({
            isOpen: true,
            title: 'Remove Member',
            message: 'Are you sure you want to remove this member from the team?',
            confirmText: 'Remove',
            confirmColor: '#ef4444',
            onConfirm: () => confirmRemoveMember(userId)
        })
    }

    const confirmRemoveMember = async (userId) => {
        if (!selectedTeamId) return
        try {
            const res = await apiFetch(`/api/team_members?team_id=${selectedTeamId}&user_id=${userId}`, { method: 'DELETE' })
            if (res.ok) fetchTeamMembers(selectedTeamId)
        } catch (e) { console.error(e) }
    }


    const handleUpdateProductivity = async (userId, newProd) => {
        try {
            const res = await apiFetch('/api/team_members', {
                method: 'POST', // Repurposing ADD as upsert/update for productivity currently
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

    const performPromotion = async (userId) => {
        try {
            const res = await apiFetch(`/api/team_members?team_id=${selectedTeamId}&user_id=${userId}`, {
                method: 'PUT',
                body: JSON.stringify({ role: 'Owner' })
            });

            if (res.ok) {
                fetchTeamMembers(selectedTeamId);
            }
        } catch (e) {
            console.error("Failed to promote member", e);
        }
    }

    const handlePromoteMember = (userId) => {
        if (!selectedTeamId) return;
        setConfirmModal({
            isOpen: true,
            title: 'Promote to Team Owner',
            message: 'Are you sure you want to promote this user to Team Owner?',
            confirmText: 'Promote',
            confirmColor: '#fbbf24', // Warning/Owner color
            onConfirm: () => performPromotion(userId)
        });
    };

    const performDemotion = async (userId) => {
        try {
            const res = await apiFetch(`/api/team_members?team_id=${selectedTeamId}&user_id=${userId}`, {
                method: 'PUT',
                body: JSON.stringify({ role: 'Member' })
            });

            if (res.ok) {
                fetchTeamMembers(selectedTeamId);
            }
        } catch (e) {
            console.error("Failed to demote member", e);
        }
    }

    const handleDemoteMember = (userId) => {
        if (!selectedTeamId) return;
        setConfirmModal({
            isOpen: true,
            title: 'Demote to Member',
            message: 'Are you sure you want to demote this Team Owner to Member?',
            confirmText: 'Demote',
            confirmColor: 'var(--text-secondary)',
            onConfirm: () => performDemotion(userId)
        });
    };


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
                        {visibleTeams.map(team => (
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
                                    onClick={(e) => { e.stopPropagation(); handleDeleteClick(team) }}
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

                            <div style={{ marginBottom: '1.5rem', padding: '1rem', background: 'rgba(255,255,255,0.05)', borderRadius: '0.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <h4 style={{ margin: 0 }}>Team Members</h4>
                                <button
                                    onClick={() => setIsAddMemberModalOpen(true)}
                                    style={{ background: 'var(--accent)', color: 'white', border: 'none', padding: '0.5rem 1rem', borderRadius: '4px', cursor: 'pointer' }}
                                >
                                    + Add Member
                                </button>
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
                                                {/* <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{tm.role}</div> */}
                                            </div>
                                        </div>

                                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                                            {tm.role === 'Owner' ? (
                                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                                                    <span style={{
                                                        fontSize: '0.8rem',
                                                        color: '#fbbf24',
                                                        fontWeight: 'bold',
                                                        border: '1px solid #fbbf24',
                                                        padding: '2px 6px',
                                                        borderRadius: '4px'
                                                    }}>
                                                        Owner
                                                    </span>
                                                    <button
                                                        onClick={() => handleDemoteMember(tm.user_id)}
                                                        title="Demote to Member"
                                                        style={{
                                                            background: 'transparent',
                                                            border: '1px solid var(--text-secondary)',
                                                            color: 'var(--text-secondary)',
                                                            padding: '2px 6px',
                                                            borderRadius: '4px',
                                                            cursor: 'pointer',
                                                            fontSize: '0.8rem'
                                                        }}
                                                    >
                                                        Demote
                                                    </button>
                                                </div>
                                            ) : (
                                                <button
                                                    onClick={() => handlePromoteMember(tm.user_id)}
                                                    title="Promote to Team Owner"
                                                    style={{
                                                        background: 'transparent',
                                                        border: '1px solid var(--text-secondary)',
                                                        color: 'var(--text-secondary)',
                                                        padding: '2px 6px',
                                                        borderRadius: '4px',
                                                        cursor: 'pointer',
                                                        fontSize: '0.8rem'
                                                    }}
                                                >
                                                    Promote
                                                </button>
                                            )}

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
                                                onClick={() => handleRemoveMemberClick(tm.user_id)}
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

            {/* Reusable Confirmation Modal */}
            <ConfirmationModal
                isOpen={confirmModal.isOpen}
                onClose={() => setConfirmModal({ ...confirmModal, isOpen: false })}
                onConfirm={confirmModal.onConfirm}
                title={confirmModal.title}
                message={confirmModal.message}
                confirmText={confirmModal.confirmText}
                confirmColor={confirmModal.confirmColor}
            />

            {/* Add Member Modal */}
            <AddMemberModal
                isOpen={isAddMemberModalOpen}
                onClose={() => setIsAddMemberModalOpen(false)}
                allUsers={users}
                teamId={selectedTeamId}
                existingMembers={teamMembers}
                onAdd={handleAddMember}
                teamName={teams.find(t => t.id === selectedTeamId)?.name}
            />
        </div>
    )
}

AdminPanel.propTypes = {
    onBack: PropTypes.func.isRequired,
    currentUser: PropTypes.object
}
