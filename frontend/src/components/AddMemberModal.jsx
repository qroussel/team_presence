import React, { useState, useMemo } from 'react';
import PropTypes from 'prop-types';

const AddMemberModal = ({ isOpen, onClose, allUsers, teamId, existingMembers, onAdd, teamName }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedUserId, setSelectedUserId] = useState('');
    const [productivity, setProductivity] = useState(100);
    const [keepOpen, setKeepOpen] = useState(false);

    // Filter users: must not be in existingMembers and must match search term
    const filteredUsers = useMemo(() => {
        return allUsers.filter(u => {
            const isMember = existingMembers.some(em => em.user_id === u.id);
            if (isMember) return false;

            const lowerSearch = searchTerm.toLowerCase();
            return (
                u.name.toLowerCase().includes(lowerSearch) ||
                u.email.toLowerCase().includes(lowerSearch)
            );
        });
    }, [allUsers, existingMembers, searchTerm]);

    if (!isOpen) return null;

    const handleSubmit = (e) => {
        e.preventDefault();
        if (!selectedUserId) return;

        onAdd(Number.parseInt(selectedUserId, 10), Number.parseInt(productivity, 10));

        // Reset fields
        setSearchTerm('');
        setSelectedUserId('');
        setProductivity(100);

        if (!keepOpen) {
            onClose();
        }
    };

    return (
        <div
            style={{
                position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
                backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1100
            }}
            onClick={onClose}
            role="presentation"
        >
            <div
                className="glass-panel"
                style={{ width: '400px', maxWidth: '90vw' }}
                onClick={e => e.stopPropagation()}
                role="dialog"
                aria-modal="true"
                aria-labelledby="modal-title"
            >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                    <h3 id="modal-title" style={{ margin: 0 }}>Add Member to {teamName}</h3>
                    <button onClick={onClose} style={{ background: 'transparent', border: 'none', color: 'var(--text-primary)', cursor: 'pointer', fontSize: '1.2rem' }}>×</button>
                </div>

                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    {/* Search / Select User */}
                    <div>
                        <label htmlFor="user-search" style={{ display: 'block', marginBottom: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Select User</label>
                        <input
                            id="user-search"
                            type="text"
                            placeholder="Search by name or email..."
                            value={searchTerm}
                            onChange={e => setSearchTerm(e.target.value)}
                            style={{
                                width: '100%',
                                padding: '0.5rem',
                                marginBottom: '0.5rem',
                                background: 'rgba(0,0,0,0.2)',
                                border: '1px solid var(--glass-border)',
                                color: 'var(--text-primary)',
                                borderRadius: '4px'
                            }}
                        />
                        <select
                            value={selectedUserId}
                            onChange={e => setSelectedUserId(e.target.value)}
                            size={5} // Show multiple options
                            style={{
                                width: '100%',
                                padding: '0.5rem',
                                background: 'rgba(0,0,0,0.2)',
                                border: '1px solid var(--glass-border)',
                                color: 'var(--text-primary)',
                                borderRadius: '4px'
                            }}
                            required
                            aria-label="Select user from list"
                        >
                            {filteredUsers.length === 0 && <option disabled>No matching users found</option>}
                            {filteredUsers.map(u => (
                                <option key={u.id} value={u.id}>
                                    {u.name} ({u.email})
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Productivity */}
                    <div>
                        <label htmlFor="productivity-input" style={{ display: 'block', marginBottom: '0.5rem', color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Productivity (%)</label>
                        <input
                            id="productivity-input"
                            type="number"
                            min="0" max="100"
                            value={productivity}
                            onChange={e => setProductivity(e.target.value)}
                            style={{
                                width: '100%',
                                padding: '0.5rem',
                                background: 'rgba(0,0,0,0.2)',
                                border: '1px solid var(--glass-border)',
                                color: 'var(--text-primary)',
                                borderRadius: '4px'
                            }}
                        />
                    </div>

                    {/* Keep Open Checkbox */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <input
                            type="checkbox"
                            id="keepOpen"
                            checked={keepOpen}
                            onChange={e => setKeepOpen(e.target.checked)}
                        />
                        <label htmlFor="keepOpen" style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Add another member</label>
                    </div>

                    <button
                        type="submit"
                        disabled={!selectedUserId}
                        style={{
                            marginTop: '0.5rem',
                            padding: '0.75rem',
                            background: 'var(--accent)',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontWeight: 'bold',
                            opacity: selectedUserId ? 1 : 0.5
                        }}
                    >
                        Add Member
                    </button>
                </form>
            </div>
        </div>
    );
};

AddMemberModal.propTypes = {
    isOpen: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
    allUsers: PropTypes.array.isRequired,
    teamId: PropTypes.number,
    existingMembers: PropTypes.array.isRequired,
    onAdd: PropTypes.func.isRequired,
    teamName: PropTypes.string
};

export default AddMemberModal;
