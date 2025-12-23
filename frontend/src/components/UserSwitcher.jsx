import React, { useEffect, useState } from 'react';
import { getSimulatedUserID, setSimulatedUserID, apiFetch } from '../utils/api';
import PropTypes from 'prop-types';

const UserSwitcher = ({ onUserChange }) => {
    const [users, setUsers] = useState([]);
    const [currentId, setCurrentId] = useState(getSimulatedUserID() || '');

    useEffect(() => {
        const loadUsers = async () => {
            try {
                // Use fetch directly to avoid infinite loops if apiFetch depends on something, 
                // but apiFetch is safe here.
                const res = await apiFetch('/api/users');
                if (res.ok) {
                    const data = await res.json();
                    setUsers(data || []);
                }
            } catch (e) {
                console.error("Failed to load users for switcher", e);
            }
        };
        loadUsers();
    }, []);

    const handleChange = (e) => {
        const newId = e.target.value;
        setCurrentId(newId);
        setSimulatedUserID(newId);
        if (onUserChange) onUserChange(newId);
        globalThis.location.reload(); // Simple reload to refresh app state with new permissions
    };

    const normalizedCurrentId = String(currentId);
    const currentUser = users.find(u => String(u.id) === normalizedCurrentId);

    return (
        <div style={{
            position: 'fixed',
            top: '1rem',
            right: '1rem',
            zIndex: 1000,
        }}>
            {/* Custom visual container */}
            <div className="glass-panel" style={{
                position: 'relative',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.5rem 1rem',
                border: '1px solid var(--accent)',
                boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
                cursor: 'pointer',
                minWidth: '200px',
                justifyContent: 'center'
            }}>
                <span style={{ fontSize: '0.9rem', fontWeight: 'bold' }}>
                    Logged in as: <span style={{ color: 'var(--primary)' }}>{currentUser ? currentUser.name : 'Guest'}</span>
                </span>

                {currentUser && (
                    <div style={{
                        width: '10px',
                        height: '10px',
                        borderRadius: '50%',
                        backgroundColor: currentUser.role === 'Admin' ? '#fbbf24' : '#34d399',
                        boxShadow: `0 0 5px ${currentUser.role === 'Admin' ? '#fbbf24' : '#34d399'}`
                    }} />
                )}

                {/* Invisible Select Overlay */}
                <select
                    value={normalizedCurrentId}
                    onChange={handleChange}
                    style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        height: '100%',
                        opacity: 0,
                        cursor: 'pointer',
                        appearance: 'none', // Remove native arrow
                    }}
                    title="Switch User"
                >
                    <option value="">Guest (Anonymous)</option>
                    {users.map(u => (
                        <option key={u.id} value={u.id}>
                            {u.name} ({u.role || 'User'})
                        </option>
                    ))}
                </select>
            </div>
        </div>
    );

};

UserSwitcher.propTypes = {
    onUserChange: PropTypes.func
};

export default UserSwitcher;
