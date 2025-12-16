import React, { useState, useEffect, useRef } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isWeekend, isSameDay, parseISO } from 'date-fns';

const TeamCalendar = ({ users }) => {
    // Default to current month
    const [startDate, setStartDate] = useState(startOfMonth(new Date()));
    const [endDate, setEndDate] = useState(endOfMonth(new Date()));

    const [presenceData, setPresenceData] = useState([]);
    const [loading, setLoading] = useState(false);
    const scrollContainerRef = useRef(null);

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
            // Optional: Auto-adjust end date if start > end?
            // For now, let user manage it, but prevent crash in days calculation
        }
    };

    const handleEndDateChange = (e) => {
        const date = parseISO(e.target.value);
        if (date.toString() !== 'Invalid Date') {
            setEndDate(date);
        }
    };

    return (
        <div className="glass-panel" style={{ display: 'flex', flexDirection: 'column' }}>
            <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
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
                {(startDate > endDate) && (
                    <span style={{ color: '#ef4444', fontSize: '0.9rem' }}>Start date must be before end date</span>
                )}

                {/* Total Working Days Display */}
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
                            if (isWeekend(day)) return total;

                            const daySum = users.reduce((userTotal, user) => {
                                const presence = getPresenceForUserAndDay(user.id, day);
                                let dailyScore = 0;

                                // AM
                                const statusAM = presence?.status_am;
                                if (statusAM !== 'off') dailyScore += 0.5;

                                // PM
                                const statusPM = presence?.status_pm;
                                if (statusPM !== 'off') dailyScore += 0.5;

                                return userTotal + dailyScore;
                            }, 0);

                            return total + daySum;
                        }, 0)
                    }
                </div>
            </div>

            <div style={{ display: 'flex' }}>
                {/* Fixed Sidebar: User Names */}
                <div style={{ flex: '0 0 200px', borderRight: '1px solid var(--glass-border)' }}>
                    <div style={{ height: '50px', display: 'flex', alignItems: 'center', fontWeight: 'bold', padding: '0 0.5rem', background: 'var(--bg-secondary)' }}>
                        Team Member
                    </div>
                    {users.map(user => (
                        <div key={user.id} style={{
                            height: '40px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.5rem',
                            padding: '0 0.5rem',
                            background: 'var(--bg-secondary)',
                            borderBottom: '1px solid transparent' // Keep alignment with grid gap
                        }}>
                            <div style={{ width: '24px', height: '24px', borderRadius: '50%', background: 'linear-gradient(135deg, #6366f1, #ec4899)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.7rem', fontWeight: 'bold' }}>
                                {user.name.charAt(0).toUpperCase()}
                            </div>
                            <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{user.name}</span>
                        </div>
                    ))}
                </div>

                {/* Scrollable Grid Area */}
                <div ref={scrollContainerRef} style={{ flex: 1, overflowX: 'auto' }}>
                    <div style={{ display: 'grid', gridTemplateColumns: `repeat(${days.length}, minmax(40px, 1fr))`, gap: '1px' }}>
                        {/* Header Row: Days */}
                        {days.map(day => (
                            <div key={day.toString()} style={{
                                height: '50px',
                                textAlign: 'center',
                                display: 'flex',
                                flexDirection: 'column',
                                justifyContent: 'center',
                                fontWeight: 'bold',
                                backgroundColor: isWeekend(day) ? 'rgba(255,255,255,0.05)' : 'var(--bg-secondary)',
                                color: isSameDay(day, new Date()) ? 'var(--accent)' : 'inherit'
                            }}>
                                <div>{format(day, 'd')}</div>
                                <div style={{ fontSize: '0.7em', fontWeight: 'normal', opacity: 0.7 }}>{format(day, 'MMM')}</div>
                            </div>
                        ))}

                        {/* Presence Grid */}
                        {users.map(user => (
                            <React.Fragment key={user.id}>
                                {days.map(day => {
                                    const presence = getPresenceForUserAndDay(user.id, day);
                                    const statusAM = presence?.status_am || '';
                                    const statusPM = presence?.status_pm || '';

                                    return (
                                        <div key={`${user.id}-${day}`} style={{
                                            height: '40px',
                                            backgroundColor: isWeekend(day) ? 'rgba(255,255,255,0.02)' : 'transparent',
                                            border: '1px solid rgba(255,255,255,0.05)',
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
