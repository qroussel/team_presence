import React, { useState, useEffect } from 'react'
import Calendar from './components/Calendar'
import TeamCalendar from './components/TeamCalendar'
import AdminPanel from './components/AdminPanel' // Import AdminPanel
import { ThemeProvider } from './context/ThemeContext'
import ThemeSwitcher from './components/ThemeSwitcher'
import UserSwitcher from './components/UserSwitcher'
import { apiFetch, getSimulatedUserID } from './utils/api'

import AddTeamModal from './components/AddTeamModal'
import UserSettingsModal from './components/UserSettingsModal'

function AppContent() {
  const [users, setUsers] = useState([])
  const [selectedUserId, setSelectedUserId] = useState(Number(getSimulatedUserID()) || null)
  const [userMemberships, setUserMemberships] = useState([]) // Stores current user's team memberships
  const [viewMode, setViewMode] = useState('personal') // 'personal' | 'team' | 'admin'
  const [settingsUser, setSettingsUser] = useState(null) // User currently being edited in settings
  const [isAddTeamModalOpen, setIsAddTeamModalOpen] = useState(false)

  // Simple form state for adding a user
  const [newUserName, setNewUserName] = useState('')
  const [newUserEmail, setNewUserEmail] = useState('')

  const [teams, setTeams] = useState([])

  const fetchUsers = async () => {
    try {
      const res = await apiFetch('/api/users')
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

  const fetchTeams = async () => {
    try {
      const res = await apiFetch('/api/teams')
      if (res.ok) {
        const data = await res.json()
        setTeams(data || [])
      }
    } catch (e) {
      console.error("Failed to fetch teams", e)
    }
  }

  useEffect(() => {
    fetchUsers()
    fetchTeams()
  }, [])

  useEffect(() => {
    if (selectedUserId) {
      apiFetch(`/api/team_members?user_id=${selectedUserId}`)
        .then(res => res.json())
        .then(data => setUserMemberships(data || []))
        .catch(console.error)
    } else {
      setUserMemberships([])
    }
  }, [selectedUserId])

  const handleDeleteUser = async (userId) => {
    try {
      const res = await apiFetch(`/api/users?id=${userId}`, { method: 'DELETE' })
      if (res.ok) {
        // Refresh users, clear selection if needed
        await fetchUsers()
        if (selectedUserId === userId) {
          setSelectedUserId(null)
        }
      } else {
        alert("Action failed - Check your permissions (Admin only)")
      }
    } catch (e) {
      console.error("Failed to delete user", e)
    }
  }

  const handleAddUser = async (e) => {
    e.preventDefault()
    if (!newUserName || !newUserEmail) return

    try {
      const res = await apiFetch('/api/users', {
        method: 'POST',
        body: JSON.stringify({
          name: newUserName,
          email: newUserEmail,
          avatar_url: '' // optional
        })
      })
      if (res.ok) {
        const newUser = await res.json()

        // Auto-add to first team (Engineering)
        if (teams.length > 0) {
          await apiFetch('/api/team_members', {
            method: 'POST',
            body: JSON.stringify({
              team_id: teams[0].id,
              user_id: newUser.id,
              productivity: 100
            })
          })
        }

        setNewUserName('')
        setNewUserEmail('')
        fetchUsers()
      } else {
        alert("Action failed - Check your permissions (Admin only)")
      }
    } catch (e) {
      console.error("Failed to add user", e)
    }
  }

  const currentUser = users.find(u => u.id === selectedUserId)

  if (viewMode === 'admin') {
    return (
      <div>
        <UserSwitcher />
        <AdminPanel
          currentUser={currentUser}
          onBack={() => {
            setViewMode('personal')
            fetchTeams() // Refresh teams when returning
            fetchUsers()
          }}
        />
      </div>
    )
  }

  return (
    <div>
      <UserSwitcher />
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
                      backgroundColor: isSelected ? 'var(--selected-item-bg)' : 'transparent',
                      transition: 'background-color 0.2s',
                      position: 'relative',
                      group: 'true' // hint for hover logic if using css, but we'll specific inline style
                    }}
                    onMouseEnter={e => e.currentTarget.querySelector('.settings-icon').style.opacity = '1'}
                    onMouseLeave={e => e.currentTarget.querySelector('.settings-icon').style.opacity = '0'}
                  >
                    <div style={{ width: '32px', height: '32px', borderRadius: '50%', background: 'linear-gradient(135deg, #6366f1, #ec4899)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.8rem', fontWeight: 'bold' }}>
                      {u.name.charAt(0).toUpperCase()}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: 500 }}>{u.name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>{u.email}</div>
                    </div>
                    <div
                      className="settings-icon"
                      onClick={(e) => {
                        e.stopPropagation(); // Prevent selection
                        setSettingsUser(u);
                      }}
                      style={{
                        opacity: 0,
                        transition: 'opacity 0.2s',
                        padding: '4px',
                        borderRadius: '4px',
                        background: 'rgba(255,255,255,0.1)'
                      }}
                      title="Settings"
                    >
                      ⚙️
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
        <TeamCalendar
          teams={teams}
          allUsers={users}
          onAddTeamClick={() => setIsAddTeamModalOpen(true)}
        />
      )}

      <UserSettingsModal
        user={settingsUser}
        isOpen={!!settingsUser}
        onClose={() => setSettingsUser(null)}
        onDeleteUser={handleDeleteUser}
        teams={teams}
      />

      <AddTeamModal
        isOpen={isAddTeamModalOpen}
        onClose={() => setIsAddTeamModalOpen(false)}
        users={users}
        onTeamCreated={() => {
          fetchTeams();
        }}
      />

      <footer style={{
        marginTop: '3rem',
        padding: '2rem',
        borderTop: '1px solid var(--glass-border)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        color: 'var(--text-secondary)'
      }}>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
          <div style={{ fontSize: '0.9rem' }}>Application Settings</div>
          <ThemeSwitcher />

          {/* Discrete Admin Access */}
          {(currentUser?.role === 'Admin' || currentUser?.role === 'admin' || userMemberships.some(m => m.role === 'Owner')) && (
            <button
              onClick={() => setViewMode('admin')}
              style={{ background: 'transparent', border: 'none', color: 'var(--text-secondary)', opacity: 0.5, fontSize: '0.8rem', cursor: 'pointer', marginTop: '1rem' }}
            >
              Admin Panel
            </button>
          )}
        </div>
      </footer>
    </div>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <AppContent />
    </ThemeProvider>
  )
}
