import React, { useState, useEffect, useRef } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isWeekend, isSameDay, parseISO } from 'date-fns';
import { isFrenchHoliday } from '../utils/holidays';

const TeamCalendar = ({ teams, allUsers, onAddTeamClick }) => {
    // Default to current month
    const [startDate, setStartDate] = useState(startOfMonth(new Date()));
    const [endDate, setEndDate] = useState(endOfMonth(new Date()));

    const [selectedTeamId, setSelectedTeamId] = useState(null);
    const [teamMembers, setTeamMembers] = useState([]);
    const [presenceData, setPresenceData] = useState([]);
    const [loading, setLoading] = useState(false);
    const scrollContainerRef = useRef(null);

    // Initial selection
    useEffect(() => {
        if (!selectedTeamId && teams.length > 0) {
            setSelectedTeamId(teams[0].id);
        }
    }, [teams]);

    // Fetch team members when team changes
    useEffect(() => {
        const fetchMembers = async () => {
            if (!selectedTeamId) return;
            try {
                const res = await fetch(`/api/team_members?team_id=${selectedTeamId}`);
                if (res.ok) {
                    const data = await res.json();
                    setTeamMembers(data || []);
                }
            } catch (e) {
                console.error("Failed to fetch team members", e);
            }
        };
        fetchMembers();
    }, [selectedTeamId]);

    // Fetch presence
    useEffect(() => {
        const fetchPresence = async () => {
            if (!startDate || !endDate) return;

            setLoading(true);
            const startStr = format(startDate, 'yyyy-MM-dd');
            const endStr = format(endDate, 'yyyy-MM-dd');

            try {
                const res = await fetch(`/api/presence?start=${startStr}&end=${endStr}`);
                if (res.ok) {
                    const data = await res.json();
                    setPresenceData(data || []);
                }
            } catch (error) {
                console.error("Failed to fetch presence", error);
            } finally {
                setLoading(false);
            }
        };

        fetchPresence();
    }, [startDate, endDate]);

    // Generate days for the selected range
    const days = (startDate && endDate && startDate <= endDate)
        ? eachDayOfInterval({ start: startDate, end: endDate })
        : [];

    const getPresenceForUserAndDay = (userId, date) => {
        const dateStr = format(date, 'yyyy-MM-dd');
        return presenceData.find(p => p.user_id === userId && p.date === dateStr);
    };

    const getBackgroundStyle = (statusAM, statusPM) => {
        const getColor = (s) => {
            switch (s) {
                case 'office': return 'rgba(16, 185, 129, 0.25)';
                case 'remote': return 'rgba(59, 130, 246, 0.25)';
                case 'off': return 'rgba(239, 68, 68, 0.25)';
                default: return null;
            }
        };

        const c1 = getColor(statusAM);
        const c2 = getColor(statusPM);
        const color1 = c1 || 'transparent';
        const color2 = c2 || 'transparent';

        const style = {
            backgroundImage: `linear-gradient(to right, ${color1} 50%, ${color2} 50%)`
        };

        if (statusAM && statusAM === statusPM) {
            if (statusAM === 'office') style.border = '1px solid #10b981';
            if (statusAM === 'remote') style.border = '1px solid #3b82f6';
            if (statusAM === 'off') style.border = '1px solid #ef4444';
        }

        return style;
    };

    const handleStartDateChange = (e) => {
        const date = parseISO(e.target.value);
        if (date.toString() !== 'Invalid Date') {
            setStartDate(date);
        }
    };

    const handleEndDateChange = (e) => {
        const date = parseISO(e.target.value);
        if (date.toString() !== 'Invalid Date') {
            setEndDate(date);
        }
    };

    const handleCreateTeam = async () => {
        const name = prompt("Enter team name:");
        if (!name) return;
        try {
            const res = await fetch('/api/teams', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name })
            });
            if (res.ok) {
                window.location.reload(); // Simple reload to refresh everything
            }
        } catch (e) {
            alert("Failed to create team");
        }
    };

    // User selection for adding to team
    const handleAddMember = async () => {
        const email = prompt("Enter user email to add to this team:");
        if (!email) return;

        const user = allUsers.find(u => u.email === email);
        if (!user) {
            alert("User not found!");
            return;
        }

        const ratioStr = prompt("Enter productivity ratio (0-100):", "100");
        const ratio = parseInt(ratioStr);
        if (isNaN(ratio) || ratio < 0 || ratio > 100) {
            alert("Invalid ratio");
            return;
        }

        try {
            const res = await fetch('/api/team_members', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    team_id: selectedTeamId,
                    user_id: user.id,
                    productivity: ratio
                })
            });
            if (res.ok) {
                // Refresh members
                const res2 = await fetch(`/api/team_members?team_id=${selectedTeamId}`);
                const data = await res2.json();
                setTeamMembers(data || []);
            }
        } catch (e) {
            alert("Failed to add member");
        }
    };

    return (
        <div className="glass-panel" style={{ display: 'flex', flexDirection: 'column' }}>
            <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap' }}>

                {/* Team Selector */}
                <select
                    value={selectedTeamId || ''}
                    onChange={e => setSelectedTeamId(Number(e.target.value))}
                    style={{ background: 'rgba(0,0,0,0.2)', border: '1px solid var(--glass-border)', color: 'white', padding: '0.5rem', borderRadius: '4px' }}
                >
                    {teams.map(t => (
                        <option key={t.id} value={t.id}>{t.name}</option>
                    ))}
                </select>
                <button onClick={onAddTeamClick} style={{ padding: '0.5rem', cursor: 'pointer' }}>+ Team</button>

                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <label style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>From:</label>
                    <input
                        type="date"
                        value={isValidDate(startDate) ? format(startDate, 'yyyy-MM-dd') : ''}
                        onChange={handleStartDateChange}
                        style={{ background: 'rgba(0,0,0,0.2)', border: '1px solid var(--glass-border)', color: 'white', padding: '0.5rem', borderRadius: '4px' }}
                    />
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <label style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>To:</label>
                    <input
                        type="date"
                        value={isValidDate(endDate) ? format(endDate, 'yyyy-MM-dd') : ''}
                        onChange={handleEndDateChange}
                        style={{ background: 'rgba(0,0,0,0.2)', border: '1px solid var(--glass-border)', color: 'white', padding: '0.5rem', borderRadius: '4px' }}
                    />
                </div>

                <div style={{
                    marginLeft: 'auto',
                    padding: '0.5rem 1rem',
                    background: 'rgba(16, 185, 129, 0.1)',
                    border: '1px solid rgba(16, 185, 129, 0.2)',
                    borderRadius: '4px',
                    color: '#10b981',
                    fontWeight: 'bold',
                    fontSize: '0.9rem'
                }}>
                    Team Working Days: {
                        days.reduce((total, day) => {
                            if (isWeekend(day) || isFrenchHoliday(day)) return total;

                            const daySum = teamMembers.reduce((userTotal, member) => {
                                const presence = getPresenceForUserAndDay(member.user_id, day);
                                let dailyScore = 0;

                                // AM
                                const statusAM = presence?.status_am;
                                if (statusAM !== 'off') dailyScore += 0.5;

                                // PM
                                const statusPM = presence?.status_pm;
                                if (statusPM !== 'off') dailyScore += 0.5;

                                // Apply productivity ratio
                                const ratio = member.productivity / 100.0;
                                return userTotal + (dailyScore * ratio);
                            }, 0);

                            return total + daySum;
                        }, 0).toFixed(1)
                    }
                </div>
            </div>

            <div style={{ display: 'flex' }}>
                {/* Fixed Sidebar: User Names */}
                <div style={{ flex: '0 0 200px', borderRight: '1px solid var(--glass-border)', display: 'flex', flexDirection: 'column', gap: '1px' }}>
                    <div style={{ height: '50px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontWeight: 'bold', padding: '0 0.5rem', background: 'var(--bg-secondary)' }}>
                        <span>Members</span>
                        <button onClick={handleAddMember} style={{ fontSize: '0.8rem', padding: '0.2rem 0.5rem' }}>+</button>
                    </div>
                    {teamMembers.map(member => (
                        <div key={member.user_id} style={{
                            height: '40px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            gap: '0.5rem',
                            padding: '0 0.5rem',
                            background: 'var(--bg-secondary)',
                            borderBottom: '1px solid transparent'
                        }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', overflow: 'hidden' }}>
                                <div style={{ width: '24px', height: '24px', borderRadius: '50%', background: 'linear-gradient(135deg, #6366f1, #ec4899)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.7rem', fontWeight: 'bold', flexShrink: 0 }}>
                                    {member.user_name ? member.user_name.charAt(0).toUpperCase() : '?'}
                                </div>
                                <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{member.user_name}</span>
                            </div>
                            <span style={{ fontSize: '0.7rem', color: 'var(--text-secondary)' }}>{member.productivity}%</span>
                        </div>
                    ))}
                </div>

                {/* Scrollable Grid Area */}
                <div ref={scrollContainerRef} style={{ flex: 1, overflowX: 'auto' }}>
                    <div style={{ display: 'grid', gridTemplateColumns: days.map(d => isWeekend(d) ? 'minmax(20px, 0.5fr)' : 'minmax(40px, 1fr)').join(' '), gap: '1px' }}>
                        {/* Header Row: Days */}
                        {days.map(day => (
                            <div key={day.toString()} style={{
                                height: '50px',
                                textAlign: 'center',
                                display: 'flex',
                                flexDirection: 'column',
                                justifyContent: 'center',
                                fontWeight: 'bold',
                                backgroundColor: isWeekend(day) || isFrenchHoliday(day) ? 'var(--header-weekend-bg)' : 'var(--bg-secondary)',
                                color: isSameDay(day, new Date()) ? 'var(--accent)' : 'inherit'
                            }}>
                                <div>{format(day, 'd')}</div>
                                <div style={{ fontSize: '0.7em', fontWeight: 'normal', opacity: 0.7 }}>{format(day, 'MMM')}</div>
                            </div>
                        ))}

                        {/* Presence Grid */}
                        {teamMembers.map(member => (
                            <React.Fragment key={member.user_id}>
                                {days.map(day => {
                                    const presence = getPresenceForUserAndDay(member.user_id, day);
                                    const statusAM = presence?.status_am || '';
                                    const statusPM = presence?.status_pm || '';

                                    return (
                                        <div key={`${member.user_id}-${day}`} style={{
                                            height: '40px',
                                            backgroundColor: isWeekend(day) || isFrenchHoliday(day) ? 'var(--weekend-bg)' : 'transparent',
                                            border: '1px solid var(--grid-border)',
                                            ...getBackgroundStyle(statusAM, statusPM)
                                        }}>
                                        </div>
                                    );
                                })}
                            </React.Fragment>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

// Helper for date input value
const isValidDate = (d) => d instanceof Date && !isNaN(d);

export default TeamCalendar;
