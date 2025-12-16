import React, { useState, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, startOfWeek, endOfWeek, eachDayOfInterval, isSameMonth, isSameDay, isWeekend } from 'date-fns';
import { isFrenchHoliday } from '../utils/holidays';
import ContextMenu from './ContextMenu';

const Calendar = ({ users, currentUser }) => {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [presenceData, setPresenceData] = useState([]);
  const [loading, setLoading] = useState(false);

  // Selection State
  const [selection, setSelection] = useState(null); // { start: Date, end: Date } or null
  const isDragSelecting = React.useRef(false);
  const dragStartDay = React.useRef(null);

  // Context Menu State
  const [contextMenu, setContextMenu] = useState(null); // { x, y, day }

  // Fetch presence data for the current month
  useEffect(() => {
    const fetchPresence = async () => {
      setLoading(true);
      const start = format(startOfMonth(currentDate), 'yyyy-MM-dd');
      const end = format(endOfMonth(currentDate), 'yyyy-MM-dd');

      try {
        const res = await fetch(`/api/presence?start=${start}&end=${end}`);
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
  }, [currentDate]);

  // Clear selection when changing month
  useEffect(() => {
    setSelection(null);
  }, [currentDate]);

  // Global mouse up to end drag
  useEffect(() => {
    const handleGlobalMouseUp = () => {
      if (isDragSelecting.current) {
        isDragSelecting.current = false;
        dragStartDay.current = null;
      }
    };
    window.addEventListener('mouseup', handleGlobalMouseUp);
    return () => window.removeEventListener('mouseup', handleGlobalMouseUp);
  }, []);

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(monthStart);
  // User wanted Monday start
  const startDate = startOfWeek(monthStart, { weekStartsOn: 1 });
  const endDate = endOfWeek(monthEnd, { weekStartsOn: 1 });

  const days = eachDayOfInterval({
    start: startDate,
    end: endDate,
  });

  const getPresenceForDay = (date) => {
    const dateStr = format(date, 'yyyy-MM-dd');
    return presenceData.filter(p => p.date === dateStr);
  };

  const updatePresence = async (userId, date, statusAM, statusPM) => {
    const dateStr = format(date, 'yyyy-MM-dd');
    try {
      const res = await fetch('/api/presence', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          date: dateStr,
          status_am: statusAM,
          status_pm: statusPM
        })
      });

      if (res.ok) {
        const newEntry = await res.json();
        setPresenceData(prev => {
          const filtered = prev.filter(p => !(p.user_id === userId && p.date === dateStr));
          return [...filtered, newEntry];
        });
      }
    } catch (error) {
      console.error("Failed to update status", error);
    }
  };

  const getNextStatus = (current) => {
    switch (current) {
      case '': return 'office';
      case 'office': return 'remote';
      case 'remote': return 'off';
      case 'off': return '';
      default: return 'office';
    }
  };

  const isDaySelected = (day) => {
    if (!selection) return false;
    const start = selection.start < selection.end ? selection.start : selection.end;
    const end = selection.start < selection.end ? selection.end : selection.start;
    return day >= start && day <= end;
  };

  // Interaction Handlers

  const handleDayMouseDown = (day) => {
    if (!currentUser) return;
    // Don't start drag if right-clicking
    isDragSelecting.current = true;
    dragStartDay.current = day;

    // Only reset selection if we are clicking OUTSIDE the current selection
    // This allows the 'Click' handler to see the full multi-day selection and trigger bulk update.
    if (!isDaySelected(day)) {
      setSelection({ start: day, end: day });
    }
  };

  const handleDayMouseEnter = (day) => {
    if (isDragSelecting.current && dragStartDay.current) {
      setSelection({
        start: dragStartDay.current,
        end: day
      });
    }
  };

  const handleDayClick = (day) => {
    if (contextMenu || !currentUser) return;

    // We check if the click target is part of the current selection range AND the selection is more than 1 day
    // If so, we bulk update.
    // If selection is just 1 day (or null), we treat it as single click.

    const isSelected = isDaySelected(day);
    const isMultiDaySelection = selection && !isSameDay(selection.start, selection.end);

    if (isSelected && isMultiDaySelection) {
      // BULK UPDATE
      // 1. Determine next status based on the CLICKED day's current status (or default)
      const dayPresence = getPresenceForDay(day);
      const myPresence = dayPresence.find(p => p.user_id === currentUser.id);
      const currentRaw = (myPresence?.status_am === myPresence?.status_pm) ? (myPresence?.status_am || '') : '';
      const nextStatus = getNextStatus(currentRaw);

      // 2. Iterate through all days in selection
      const start = selection.start < selection.end ? selection.start : selection.end;
      const end = selection.start < selection.end ? selection.end : selection.start;

      const daysToUpdate = eachDayOfInterval({ start, end });

      daysToUpdate.forEach(d => {
        // Exclude weekends from bulk change
        if (!isWeekend(d) && !isFrenchHoliday(d)) {
          updatePresence(currentUser.id, d, nextStatus, nextStatus);
        }
      });

      // Optionally clear selection after bulk update?
      // User didn't specify, but often it's nice to clear or keep.
      // Keeping it allows cycling status multiple times (Office -> Remote -> Off).
      // So we do NOT clear selection here.

    } else {
      // SINGLE UPDATE
      // If we clicked outside or single day, we might want to clear multi-selection
      // But handleDayMouseDown already set selection to just this day if it was a fresh click.
      // So effectively we are just updating this one day.

      const dayPresence = getPresenceForDay(day);
      const myPresence = dayPresence.find(p => p.user_id === currentUser.id);
      const currentRaw = (myPresence?.status_am === myPresence?.status_pm) ? (myPresence?.status_am || '') : '';
      const nextStatus = getNextStatus(currentRaw);

      updatePresence(currentUser.id, day, nextStatus, nextStatus);

      // Reset selection to just this day
      setSelection(null);
    }
  };

  const handleContextMenu = (e, day) => {
    e.preventDefault();

    if (!currentUser) return;

    let x = e.clientX;
    let y = e.clientY;

    const estimatedWidth = 220;
    if (x + estimatedWidth > window.innerWidth) {
      x -= estimatedWidth;
    }

    setContextMenu({
      x,
      y,
      day
    });
  };

  const handleContextMenuSelect = (period, status) => {
    if (!contextMenu || !currentUser) return;
    const { day } = contextMenu;
    const dayPresence = getPresenceForDay(day);
    const myPresence = dayPresence.find(p => p.user_id === currentUser.id);

    let newAM = myPresence?.status_am || '';
    let newPM = myPresence?.status_pm || '';

    if (period === 'AM') newAM = status;
    if (period === 'PM') newPM = status;

    updatePresence(currentUser.id, day, newAM, newPM);
    setContextMenu(null);
  };

  // Helper to get background style
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

    // Restore border if single status (AM == PM)
    if (statusAM && statusAM === statusPM) {
      if (statusAM === 'office') style.border = '1px solid #10b981';
      if (statusAM === 'remote') style.border = '1px solid #3b82f6';
      if (statusAM === 'off') style.border = '1px solid #ef4444';
    }

    return style;
  };

  return (
    <div className="glass-panel" onMouseLeave={() => { isDragSelecting.current = false; }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <button onClick={() => setCurrentDate(new Date(currentDate.setMonth(currentDate.getMonth() - 1)))}>&lt; Prev</button>
        <h2 style={{ margin: 0 }}>
          {format(currentDate, 'MMMM yyyy')}
          {currentUser && <span style={{ fontSize: '0.8em', marginLeft: '1rem', opacity: 0.7, fontWeight: 'normal' }}>- {currentUser.name}</span>}
        </h2>
        <button onClick={() => setCurrentDate(new Date(currentDate.setMonth(currentDate.getMonth() + 1)))}>Next &gt;</button>
      </div>

      <div className="calendar-grid" style={{ userSelect: 'none' }}>
        {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(day => (
          <div key={day} className="header-cell">{day}</div>
        ))}

        {days.map(day => {
          const dayPresence = getPresenceForDay(day);
          const myPresence = currentUser ? dayPresence.find(p => p.user_id === currentUser.id) : null;
          const statusAM = myPresence?.status_am || '';
          const statusPM = myPresence?.status_pm || '';
          const selected = isDaySelected(day);

          const isHoliday = isFrenchHoliday(day);

          return (
            <div
              key={day.toString()}
              className={`calendar-cell ${isWeekend(day) || isHoliday ? 'weekend-cell' : ''}`}
              style={{
                opacity: isSameMonth(day, monthStart) ? 1 : 0.5,
                cursor: currentUser ? 'pointer' : 'default',
                border: selected
                  ? '2px solid white'
                  : (statusAM || statusPM) ? undefined : '1px solid transparent',
                transform: selected ? 'scale(0.98)' : 'scale(1)',
                transition: 'transform 0.1s',
                boxShadow: selected ? '0 0 10px rgba(255,255,255,0.1)' : 'none',
                ...getBackgroundStyle(statusAM, statusPM)
              }}
              onMouseDown={() => handleDayMouseDown(day)}
              onMouseEnter={() => handleDayMouseEnter(day)}
              onClick={() => handleDayClick(day)}
              onContextMenu={(e) => handleContextMenu(e, day)}
            >
              <div style={{ fontWeight: 600, marginBottom: '0.5rem', color: isSameDay(day, new Date()) ? 'var(--accent)' : 'inherit' }}>
                {format(day, 'd')}
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                {(statusAM || statusPM) && (
                  <div style={{ display: 'flex', gap: '2px', marginBottom: '4px' }}>
                    <div className={`presence-indicator status-${statusAM}`} style={{ width: '6px', height: '6px' }}></div>
                    <div className={`presence-indicator status-${statusPM}`} style={{ width: '6px', height: '6px' }}></div>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>

      <div style={{ marginTop: '1.5rem' }}>
        <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center', justifyContent: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}><span className="presence-indicator status-office"></span> Office (Green)</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}><span className="presence-indicator status-remote"></span> Remote (Blue)</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}><span className="presence-indicator status-off"></span> Off (Red)</div>
        </div>
        <div style={{ marginTop: '0.5rem', fontSize: '0.8rem', color: 'var(--text-secondary)', textAlign: 'center' }}>
          * Drag to select multiple days. Click to cycle status (Office &rarr; Remote &rarr; Off &rarr; None).
        </div>
        <div style={{ marginTop: '0.2rem', fontSize: '0.8rem', color: 'var(--text-secondary)', textAlign: 'center' }}>
          * Right-click for AM/PM split.
        </div>

      </div>

      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          onClose={() => setContextMenu(null)}
          onSelect={handleContextMenuSelect}
        />
      )}
    </div>
  );
};

export default Calendar;
