import React, { useState, useEffect } from 'react'
import Calendar from './components/Calendar'
import TeamCalendar from './components/TeamCalendar'

function App() {
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState(null)
  const [viewMode, setViewMode] = useState('personal') // 'personal' | 'team'

  // Simple form state for adding a user
  const [newUserName, setNewUserName] = useState('')
  const [newUserEmail, setNewUserEmail] = useState('')

  const fetchUsers = async () => {
    try {
      const res = await fetch('/api/users')
      if (res.ok) {
        const data = await res.json()
        setUsers(data || [])
        // Default to first user if none selected and users exist
        if (!selectedUserId && data && data.length > 0) {
          setSelectedUserId(data[0].id)
        }
      }
    } catch (e) {
      console.error("Failed to fetch users", e)
    }
  }

  useEffect(() => {
    fetchUsers()
  }, [])

  const handleAddUser = async (e) => {
    e.preventDefault()
    if (!newUserName || !newUserEmail) return

    try {
      const res = await fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newUserName,
          email: newUserEmail,
          avatar_url: '' // optional
        })
      })
      if (res.ok) {
        setNewUserName('')
        setNewUserEmail('')
        fetchUsers()
      }
    } catch (e) {
      console.error("Failed to add user", e)
    }
  }

  const currentUser = users.find(u => u.id === selectedUserId)

  return (
    <div>
      <header style={{ marginBottom: '2rem', textAlign: 'center', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
        <div>
          <h1>Team Presence</h1>
          <p style={{ color: 'var(--text-secondary)' }}>Plan your team's availability efficiently.</p>
        </div>

        <div style={{ display: 'flex', background: 'rgba(255,255,255,0.1)', padding: '4px', borderRadius: '8px' }}>
          <button
            onClick={() => setViewMode('personal')}
            style={{
              background: viewMode === 'personal' ? 'var(--accent)' : 'transparent',
              color: viewMode === 'personal' ? 'white' : 'var(--text-secondary)',
              border: 'none',
              padding: '0.5rem 1rem',
              borderRadius: '6px',
              cursor: 'pointer',
              fontWeight: 500
            }}
          >
            Personal View
          </button>
          <button
            onClick={() => setViewMode('team')}
            style={{
              background: viewMode === 'team' ? 'var(--accent)' : 'transparent',
              color: viewMode === 'team' ? 'white' : 'var(--text-secondary)',
              border: 'none',
              padding: '0.5rem 1rem',
              borderRadius: '6px',
              cursor: 'pointer',
              fontWeight: 500
            }}
          >
            Team View
          </button>
        </div>
      </header>

      {viewMode === 'personal' ? (
        <div style={{ display: 'grid', gridTemplateColumns: 'minmax(250px, 1fr) 3fr', gap: '2rem' }}>

          {/* Sidebar / User List */}
          <div className="glass-panel" style={{ height: 'fit-content' }}>
            <h3>Team Members</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginBottom: '1.5rem' }}>
              {users.length === 0 && <span style={{ color: 'var(--text-secondary)' }}>No users found.</span>}
              {users.map(u => {
                const isSelected = u.id === selectedUserId
                return (
                  <div
                    key={u.id}
                    onClick={() => setSelectedUserId(u.id)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      cursor: 'pointer',
                      padding: '0.5rem',
                      borderRadius: '0.5rem',
                      backgroundColor: isSelected ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
                      transition: 'background-color 0.2s'
                    }}
                  >
                    <div style={{ width: '32px', height: '32px', borderRadius: '50%', background: 'linear-gradient(135deg, #6366f1, #ec4899)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.8rem', fontWeight: 'bold' }}>
                      {u.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div style={{ fontWeight: 500 }}>{u.name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{u.email}</div>
                    </div>
                  </div>
                )
              })}
            </div>

            <div style={{ borderTop: '1px solid var(--glass-border)', paddingTop: '1rem' }}>
              <h4 style={{ marginTop: 0, marginBottom: '0.5rem' }}>Add Member</h4>
              <form onSubmit={handleAddUser} style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <input
                  placeholder="Name"
                  value={newUserName}
                  onChange={e => setNewUserName(e.target.value)}
                />
                <input
                  placeholder="Email"
                  value={newUserEmail}
                  onChange={e => setNewUserEmail(e.target.value)}
                />
                <button type="submit">Add</button>
              </form>
            </div>
          </div>

          {/* Main Calendar Area */}
          <div>
            <Calendar users={users} currentUser={currentUser} />
          </div>

        </div>
      ) : (
        <TeamCalendar users={users} />
      )}
    </div>
  )
}

export default App
