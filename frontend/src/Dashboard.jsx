import React, { useState, useEffect } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, Legend
} from 'recharts';
import {
  DndContext, closestCenter, MouseSensor, useSensor, useSensors,
  PointerSensor, DragOverlay, useDroppable
} from '@dnd-kit/core';
import {
  SortableContext, useSortable, verticalListSortingStrategy
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

const API = 'http://localhost:8080';

const TYPE_COLORS = {
  Core: '#6366F1', Elective: '#22D3EE', Honors: '#F59E0B',
  Minors: '#10B981', 'Open Elective': '#EC4899',
};

// ─── Draggable Course Card ──────────────────────────────────────────
function CourseCard({ slot, onDrilldown, isConflictState, isMismatchState, initialSlots }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: slot.id });

  if (isDragging) {
    return <div ref={setNodeRef} style={{ opacity: 0.3 }} className="course-card" />;
  }

  const style = { transform: CSS.Transform.toString(transform), transition };
  const isConflict = isConflictState;
  const isMismatch = isMismatchState;
  const isBoth = isConflict && isMismatch;
  
  // A card is "manual" if it has been moved from its starting position
  const original = initialSlots.find(s => s.id === slot.id);
  const isManual = original && (slot.day_number !== original.day_number || slot.session !== original.session);

  const className = `course-card ${isBoth ? 'conflict-striped' : isConflict ? 'conflict-glow' : isMismatch ? 'mismatch-glow' : isManual ? 'manual-glow' : ''}`;

  const codes = (slot.course_codes?.split(',') ?? [])
    .map(c => c.replace(/[^a-zA-Z0-9]/g, ''))
    .filter((v, i, a) => v && a.indexOf(v) === i);
  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}
      className={className}
      onClick={() => onDrilldown?.(slot)}>
      <div className="course-card-codes" style={{ 
        color: isConflictState ? '#B91C1C' : isMismatchState ? '#92400E' : 'var(--primary)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <span>{codes.join(' / ')}</span>
        <span style={{ 
          fontSize: 10, 
          background: isConflictState ? '#FEE2E2' : isMismatchState ? '#FEF3C7' : '#F1F5F9', 
          color: isConflictState ? '#B91C1C' : isMismatchState ? '#92400E' : '#475569',
          padding: '2px 6px',
          borderRadius: 4,
          fontWeight: 800
        }}>
          S{(slot.semesters || "").split(',').filter((v, i, a) => v && a.indexOf(v) === i).join(', ')}
        </span>
      </div>
      <div className="course-card-name">{slot.course_name}</div>
      <div className="course-card-meta">
        <span className="course-card-dept">{slot.departments?.split(',').filter((v, i, a) => a.indexOf(v) === i).join(', ')}</span>
        <span className="course-card-strength">👥 {slot.strength}</span>
      </div>
    </div>
  );
}

function DroppableSession({ id, children }) {
  const { setNodeRef, isOver } = useDroppable({ id });
  return <div ref={setNodeRef} className="tt-pdf-courses-col" style={{ background: isOver ? '#F8FAFC' : 'transparent' }}>{children}</div>;
}

// ─── Metric Card ────────────────────────────────────────────────────
function MetricCard({ label, value, icon, color, sub }) {
  return (
    <div className="metric-card" style={{ borderTop: `3px solid ${color}` }}>
      <div className="metric-icon" style={{ background: color + '22', color }}>{icon}</div>
      <div className="metric-body">
        <div className="metric-value">{value ?? '—'}</div>
        <div className="metric-label">{label}</div>
        {sub && <div className="metric-sub">{sub}</div>}
      </div>
    </div>
  );
}

function SessionDetailsModal({ timetableId, day, period, onClose }) {
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    fetch(`${API}/api/timetable/${timetableId}/session/${day}/${period}/students`)
      .then(r => r.json())
      .then(d => {
        if (d.success) setStudents(d.students || []);
        setLoading(false);
      });
  }, [timetableId, day, period]);

  const filtered = students.filter(s =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    s.roll.toLowerCase().includes(search.toLowerCase()) ||
    s.course_name.toLowerCase().includes(search.toLowerCase()) ||
    (s.course_code && s.course_code.toLowerCase().includes(search.toLowerCase())) ||
    (s.type && s.type.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div className="modal-overlay" onClick={onClose} style={{ zIndex: 9999 }}>
      <div className="drilldown-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '900px', width: '95%', maxHeight: '85vh', display: 'flex', flexDirection: 'column' }}>
        <div className="modal-header" style={{ flexShrink: 0 }}>
          <h2 style={{ fontSize: 16 }}>📋 Session Student List: Day {day} ({period})</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', fontSize: 24, cursor: 'pointer' }}>&times;</button>
        </div>
        <div style={{ padding: 24, flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <input
              type="text"
              placeholder="Search by student name, roll number, or course..."
              className="form-input"
              style={{ maxWidth: '400px', fontSize: 13 }}
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
            <div style={{ fontSize: 12, color: 'var(--text-secondary)', fontWeight: 600 }}>
              {filtered.length} Students Scheduled
            </div>
          </div>

          {students.some(s => s.is_clashing) && (
            <div style={{ 
              background: '#FEF2F2', 
              border: '1px solid #FCA5A5', 
              padding: '10px 16px', 
              borderRadius: 8, 
              marginBottom: 16,
              color: '#991B1B',
              fontSize: 13,
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              fontWeight: 600
            }}>
              <span>⚠️ ATTENTION: {students.filter(s => s.is_clashing).length} student(s) in this session have multiple exams scheduled at once!</span>
            </div>
          )}

          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-secondary)' }}>Loading student records...</div>
          ) : (
            <div style={{ flex: 1, overflowY: 'auto', border: '1px solid var(--border)', borderRadius: 8 }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
                <thead style={{ background: '#F8FAFC', position: 'sticky', top: 0, zIndex: 10 }}>
                  <tr>
                    <th style={{ textAlign: 'left', padding: '12px 16px', borderBottom: '2px solid var(--border)' }}>Roll Number</th>
                    <th style={{ textAlign: 'left', padding: '12px 16px', borderBottom: '2px solid var(--border)' }}>Student Name</th>
                    <th style={{ textAlign: 'left', padding: '12px 16px', borderBottom: '2px solid var(--border)' }}>Type</th>
                    <th style={{ textAlign: 'left', padding: '12px 16px', borderBottom: '2px solid var(--border)' }}>Subject</th>
                    <th style={{ textAlign: 'left', padding: '12px 16px', borderBottom: '2px solid var(--border)' }}>Department</th>
                  </tr>
                </thead>
                <tbody>
                   {filtered.length === 0 ? (
                    <tr><td colSpan="5" style={{ padding: 40, textAlign: 'center', color: 'var(--text-secondary)' }}>No records found matching your search.</td></tr>
                  ) : filtered.map((s, idx) => (
                    <tr key={idx} style={{ 
                      borderBottom: s.is_clashing ? '2px solid #F87171' : '1px solid var(--border)',
                      background: s.is_clashing ? '#FEE2E2' : 'transparent',
                      color: s.is_clashing ? '#991B1B' : 'inherit',
                      transition: 'all 0.2s',
                    }}>
                      <td style={{ padding: '12px 16px', fontWeight: 700, color: s.is_clashing ? '#B91C1C' : 'var(--primary)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          {s.is_clashing && <span className="clash-badge">CLASH</span>}
                          {s.roll}
                        </div>
                      </td>
                      <td style={{ padding: '12px 16px', fontWeight: s.is_clashing ? 800 : 600 }}>{s.name}</td>
                      <td style={{ padding: '12px 16px' }}>
                         <span style={{ 
                           padding: '2px 8px', 
                           borderRadius: 12, 
                           fontSize: 10, 
                           fontWeight: 800,
                           background: s.type === 'Regular' ? '#DBEAFE' : '#FFEDD5',
                           color: s.type === 'Regular' ? '#1E40AF' : '#9A3412',
                           textTransform: 'uppercase'
                         }}>
                           {s.type}
                         </span>
                       </td>
                      <td style={{ padding: '12px 16px' }}>
                        <div style={{ fontSize: 11, fontWeight: 700, opacity: 0.8 }}>{s.course_code}</div>
                        <div style={{ fontWeight: s.is_clashing ? 800 : 400 }}>{s.course_name}</div>
                      </td>
                      <td style={{ padding: '12px 16px', fontSize: 12, fontWeight: s.is_clashing ? 700 : 400 }}>{s.dept}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
        <div className="modal-footer" style={{ padding: '16px 24px', background: '#F8FAFC', borderTop: '1px solid var(--border)', textAlign: 'right', flexShrink: 0 }}>
          <button className="btn-primary" onClick={onClose} style={{ padding: '8px 24px' }}>Close</button>
        </div>
      </div>
    </div>
  );
}

// ─── Main Dashboard ─────────────────────────────────────────────────
export default function Dashboard({ timetableId, onApproved, onRejected, onViewChange }) {
  const [timetable, setTimetable] = useState(null);
  const [slots, setSlots] = useState([]);
  const [initialSlots, setInitialSlots] = useState([]);
  const [history, setHistory] = useState([]);
  const [stats, setStats] = useState(null);
  const [drilldown, setDrilldown] = useState(null);
  const [warningModal, setWarningModal] = useState(null);
  const [activeId, setActiveId] = useState(null);
  const [overSession, setOverSession] = useState(null);
  const [sessionListView, setSessionListView] = useState(null); // { day, period }
  const [approving, setApproving] = useState(false);
  const [approved, setApproved] = useState(false);
  const [clashingIds, setClashingIds] = useState({}); // { slotId: true }
  
  // Filtering states
  const [filterSearch, setFilterSearch] = useState('');
  const [filterDayType, setFilterDayType] = useState('all'); // all, odd, even
  const [filterSession, setFilterSession] = useState('all'); // all, FN, AN

  const activeSlot = slots.find(s => s.id === activeId);

  // Real-time conflict check for UI feedback
  const isConflict = (sessionStr) => {
    if (!activeSlot || !sessionStr) return false;

    // Check if sessionStr is a card ID or a session ID
    let targetDay, targetSess;
    if (String(sessionStr).includes('-')) {
      const [d, s] = String(sessionStr).split('-');
      targetDay = parseInt(d);
      targetSess = s;
    } else {
      const targetCard = slots.find(s => s.id === sessionStr);
      if (!targetCard) return false;
      targetDay = targetCard.day_number;
      targetSess = targetCard.session;
    }

    // Don't flag conflict if we're hovering over our original slot
    if (targetDay === activeSlot.day_number && targetSess === activeSlot.session) return false;

    const otherCoursesInSession = slots.filter(s =>
      s.day_number === targetDay &&
      s.session === targetSess &&
      s.id !== activeId
    );

    const myDepts = activeSlot.departments?.split(',').map(d => d.trim()) || [];
    const mySems = (activeSlot.semesters || "").split(',').map(s => s.trim()).filter(s => s !== "");

    for (const other of otherCoursesInSession) {
      const otherDepts = other.departments?.split(',').map(d => d.trim()) || [];
      const otherSems = (other.semesters || "").split(',').map(s => s.trim()).filter(s => s !== "");
      
      const deptMatch = myDepts.some(d => otherDepts.includes(d));
      const semMatch = mySems.some(s => otherSems.includes(s));

      if (deptMatch && semMatch) return true;
    }
    return false;
  };

  const isMismatch = (sessionStr) => {
    if (!activeSlot || !sessionStr || !timetable?.day_config) return false;

    let targetDay, targetSess;
    if (String(sessionStr).includes('-')) {
      const [d, s] = String(sessionStr).split('-');
      targetDay = parseInt(d);
      targetSess = s;
    } else {
      const targetCard = slots.find(s => s.id === sessionStr);
      if (!targetCard) return false;
      targetDay = targetCard.day_number;
      targetSess = targetCard.session;
    }

    const dayType = (targetDay % 2 !== 0) ? 'Odd' : 'Even';
    const configKey = `${dayType}${targetSess}`;
    const allowedSems = (timetable.day_config?.[configKey] || []).map(String);
    
    // Add Extra Sem for Mode 5
    if (timetable.day_config?.mode === 5 && timetable.day_config?.extra_sem) {
      allowedSems.push(String(timetable.day_config.extra_sem));
    }

    const mySems = (activeSlot.semesters || "").split(',').map(s => s.trim()).filter(s => s !== "");
    const mismatch = mySems.some(s => !allowedSems.includes(s));
    return mismatch;
  };

  const getSlotConflict = (id) => {
    if (!clashingIds) return false;
    return !!clashingIds[id] || !!clashingIds[String(id)];
  };

  const getSlotMismatch = (chkSlot) => {
    if (!timetable?.day_config) return false;
    const dayType = (chkSlot.day_number % 2 !== 0) ? 'Odd' : 'Even';
    const configKey = `${dayType}${chkSlot.session}`;
    const allowedSems = (timetable.day_config[configKey] || []).map(String);
    
    // Add Extra Sem for Mode 5
    if (timetable.day_config?.mode === 5 && timetable.day_config?.extra_sem) {
      allowedSems.push(String(timetable.day_config.extra_sem));
    }

    const mySems = (chkSlot.semesters || "").split(',').map(s => s.trim()).filter(Boolean);
    return mySems.some(s => !allowedSems.includes(s));
  };

  const API = 'http://localhost:8080';

  // Load timetable metadata + slots
  const loadTimetable = (id) => {
    if (!id) return;
    fetch(`${API}/api/timetable/${id}`)
      .then(r => r.json())
      .then(d => {
        if (d.success) {
          setSlots(d.slots || []);
          console.log("DEBUG: Clashing IDs from Backend:", d.clashing_ids);
          setClashingIds(d.clashing_ids || {});
          if (initialSlots.length === 0) setInitialSlots(JSON.parse(JSON.stringify(d.slots || [])));

          let meta = d.timetable || {};
          if (typeof meta.day_config === 'string') {
            try { meta.day_config = JSON.parse(meta.day_config); } catch (e) { }
          }
          setTimetable(meta);
        }
      });
  };

  useEffect(() => {
    if (timetableId) {
      loadTimetable(timetableId);
    } else {
      fetch(`${API}/api/timetables?status=pending`)
        .then(r => r.json())
        .then(d => {
          if (d.success && d.timetables?.length > 0) {
            loadTimetable(d.timetables[0].id);
          }
        });
    }
  }, [timetableId]);

  // Load stats once timetable metadata loaded
  useEffect(() => {
    if (!timetable) return;

    // Extract semesters from day_config to filter stats
    const config = timetable.day_config || {};
    const sems = [];
    Object.values(config).forEach(val => {
      if (Array.isArray(val)) sems.push(...val); // For older flat configs or simple modes
      if (val.FN && Array.isArray(val.FN)) sems.push(...val.FN);
      if (val.AN && Array.isArray(val.AN)) sems.push(...val.AN);
    });
    const uniqueSems = [...new Set(sems)].join(',');

    fetch(`${API}/api/dashboard/stats?timetable_id=${timetable.id}&regulation_id=${timetable.regulation_id}&upload_type=${timetable.upload_type}&semesters=${uniqueSems}`)
      .then(r => r.json())
      .then(d => { if (d.success) setStats(d); });
  }, [timetable]);

  if (!timetableId && !timetable) {
    return (
      <main className="main-content dashboard-page">
        <div className="page-header">
          <h1>📊 Analytics Dashboard</h1>
          <p>Generate a timetable first to review and approve it here.</p>
        </div>
        <div style={{ textAlign: 'center', padding: '80px 0', color: 'var(--text-secondary)' }}>
          <div style={{ fontSize: 64, marginBottom: 16 }}>📅</div>
          <p style={{ fontSize: 16 }}>No timetable loaded. Go to <strong>Create Timetable</strong> and click Generate.</p>
        </div>
      </main>
    );
  }

  // Group slots by day + session
  const grouped = {};
  for (const slot of slots) {
    if (!grouped[slot.day_number]) grouped[slot.day_number] = { FN: [], AN: [] };
    grouped[slot.day_number][slot.session].push(slot);
  }
  const days = Object.keys(grouped).map(Number).sort((a, b) => a - b);
  const totalDays = days.length;

  // Chart data
  const pieData = stats ? Object.entries(stats.type_distribution || {}).map(([name, value]) => ({ name, value })) : [];
  const barData = stats ? (stats.dept_stats || []).map(d => ({ dept: d.department, count: d.count })) : [];
  const maxSemCount = Math.max(...Object.values(stats?.sem_map || {}), 1);

  // Manual Reschedule Handlers
  const handleUndo = async () => {
    if (history.length === 0) return;
    const last = history[history.length - 1];

    const res = await fetch(`${API}/api/timetable/${timetableId}/slot/${last.slotId}/move`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ new_day: last.fromDay, new_session: last.fromSess }),
    });

    if (res.ok) {
      setHistory(prev => prev.slice(0, -1));
      loadTimetable(timetableId || (timetable && timetable.id));
    }
  };

  const handleUndoAll = async () => {
    if (!initialSlots || !history.length) return;
    if (confirm("Undo ALL manual changes and revert to original generated state?")) {
      // Revert in DB
      for (const original of initialSlots) {
        const current = slots.find(s => s.id === original.id);
        if (current && (current.day_number !== original.day_number || current.session !== original.session)) {
          await fetch(`${API}/api/timetable/${timetableId || (timetable && timetable.id)}/slot/${original.id}/move`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ new_day: original.day_number, new_session: original.session }),
          });
        }
      }
      setHistory([]);
      loadTimetable(timetableId || (timetable && timetable.id));
    }
  };

  // Drag end
  const handleDragEnd = async ({ active, over }) => {
    setActiveId(null);
    setOverSession(null);
    if (!over || active.id === over.id) return;

    const fromSlot = slots.find(s => s.id === active.id);
    let toSlot = slots.find(s => s.id === over.id);

    if (!toSlot && String(over.id).includes('-')) {
      const [d, s] = String(over.id).split('-');
      toSlot = { day_number: parseInt(d), session: s };
    }

    if (!fromSlot || !toSlot) return;

    if (fromSlot.day_number !== toSlot.day_number || fromSlot.session !== toSlot.session) {
      const conflict = isConflict(`${toSlot.day_number}-${toSlot.session}`);
      const mismatch = isMismatch(`${toSlot.day_number}-${toSlot.session}`);

      // Check for clashing students (DEEP CHECK)
      const checkRes = await fetch(`${API}/api/timetable/${timetableId || (timetable && timetable.id)}/slot/${fromSlot.id}/check-conflicts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ new_day: toSlot.day_number, new_session: toSlot.session }),
      });
      const checkData = await checkRes.json();
      const clashingStudents = checkData.conflicts || [];

      setWarningModal({
        fromSlot, toSlot,
        isConflict: conflict || clashingStudents.length > 0,
        clashingStudents: clashingStudents,
        isMismatch: mismatch,
        onConfirm: async () => {
          const res = await fetch(`${API}/api/timetable/${timetableId || (timetable && timetable.id)}/slot/${fromSlot.id}/move`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ new_day: toSlot.day_number, new_session: toSlot.session }),
          });
          if (res.ok) {
            setHistory(prev => [...prev, { slotId: fromSlot.id, fromDay: fromSlot.day_number, fromSess: fromSlot.session }]);
            // Refresh ALL data from backend to ensure positions, counts, and glows are 100% accurate
            loadTimetable(timetableId || (timetable && timetable.id));
          }
          setWarningModal(null);
        }
      });
    }
  };

  // Approve
  const handleApprove = async () => {
    const id = timetableId || timetable?.id;
    if (!id) return;
    setApproving(true);
    const res = await fetch(`${API}/api/timetable/${id}/approve`, { method: 'PUT' });
    const data = await res.json();
    if (data.success) {
      setApproved(true);
      // Don't immediately navigate away, allow user to see the "Approved" state for a moment
      setTimeout(() => {
        if (onApproved) onApproved();
      }, 1500);
    } else {
      alert('Approval failed: ' + data.error);
    }
    setApproving(false);
  };

  const handleReject = async () => {
    const id = timetableId || timetable?.id;
    if (!id) return;
    if (!window.confirm('Discard this draft? All manual adjustments will be lost.')) return;
    const res = await fetch(`${API}/api/timetable/${id}`, { method: 'DELETE' });
    if (res.ok) {
      if (onRejected) onRejected();
      onViewChange('create');
    }
  };

  return (
    <main className="main-content dashboard-page">
      {/* ── Header ── */}
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <h1>📊 {timetable?.name || 'Dashboard'}</h1>
          <p style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span>Regulation {timetable?.regulation_id}</span>
            {timetable?.day_config && (
              <span style={{ background: '#EEF2FF', color: '#4F46E5', padding: '2px 8px', borderRadius: 4, fontSize: 11, fontWeight: 600 }}>
                🔍 Filtering: Semesters {[...new Set(Object.values(timetable.day_config).flatMap(v => [...(v.FN || []), ...(v.AN || [])]))].sort().join(', ')}
              </span>
            )}
          </p>
        </div>
        <div style={{ display: 'flex', gap: 12, paddingTop: 8 }}>
          {history.length > 0 && !approved && (
            <>
              <button className="btn-secondary" onClick={handleUndo} style={{ padding: '8px 16px', fontSize: 13, background: '#F8FAFC' }}>
                ↩️ Undo Move ({history.length})
              </button>
              <button className="btn-secondary" onClick={handleUndoAll} style={{ padding: '8px 16px', fontSize: 13, color: '#E11D48', borderColor: '#FDA4AF' }}>
                🗑️ Undo All
              </button>
            </>
          )}
          {approved ? (
            <div style={{ background: '#D1FAE5', color: '#065F46', padding: '10px 20px', borderRadius: 8, fontWeight: 700, fontSize: 14 }}>
              ✅ Approved! View in Time Tables sidebar.
            </div>
          ) : (
            <>
              <button
                className="btn-primary"
                style={{
                  background: 'linear-gradient(135deg, #EF4444, #DC2626)',
                  boxShadow: '0 4px 14px rgba(239, 68, 68, 0.35)',
                  border: 'none',
                  color: 'white',
                  transition: 'all 0.2s ease'
                }}
                onMouseOver={e => e.currentTarget.style.transform = 'translateY(-2px)'}
                onMouseOut={e => e.currentTarget.style.transform = 'translateY(0)'}
                onClick={handleReject}
              >
                ✖ Discard Draft
              </button>
              <button className="btn-primary" style={{ background: 'linear-gradient(135deg, #10B981, #059669)', boxShadow: '0 4px 14px rgba(16,185,129,0.35)' }}
                onClick={handleApprove} disabled={approving}>
                {approving ? 'Approving...' : '✅ Approve & Generate'}
              </button>
            </>
          )}
        </div>
      </div>

      {/* ── Configuration Summary Bar ── */}
      {timetable?.day_config && (
        <div style={{ 
          background: 'white', 
          border: '1px solid var(--border)', 
          borderRadius: 12, 
          padding: '12px 20px', 
          marginBottom: 20,
          display: 'flex',
          alignItems: 'center',
          gap: 16,
          boxShadow: '0 2px 4px rgba(0,0,0,0.02)'
        }}>
          <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
            ⚙️ CONFIGURATION:
          </span>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            {(() => {
              const cfg = timetable.day_config;
              const mode = cfg.mode || 4;
              let items = [];

              if (mode === 1) {
                // All sessions are the same
                items.push({ label: 'Daily', sems: (cfg.OddFN || []) });
              } else if (mode === 2) {
                // All odd/even are the same, but FN/AN split
                items.push({ label: 'Daily FN', sems: (cfg.OddFN || []) });
                items.push({ label: 'Daily AN', sems: (cfg.OddAN || []) });
              } else {
                // Mode 4 or 5
                if(cfg.OddFN && cfg.OddFN.length > 0) items.push({ label: 'Odd FN', sems: cfg.OddFN });
                if(cfg.OddAN && cfg.OddAN.length > 0) items.push({ label: 'Odd AN', sems: cfg.OddAN });
                if(cfg.EvenFN && cfg.EvenFN.length > 0) items.push({ label: 'Even FN', sems: cfg.EvenFN });
                if(cfg.EvenAN && cfg.EvenAN.length > 0) items.push({ label: 'Even AN', sems: cfg.EvenAN });
                
                if (mode === 5 && cfg.extra_sem) {
                  items.push({ label: 'Extra', sems: [cfg.extra_sem], isExtra: true });
                }
              }

              return items.map((item, idx) => (
                <div key={idx} style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: 6,
                  background: item.isExtra ? 'var(--primary-light)' : '#F8FAFC',
                  padding: '4px 10px',
                  borderRadius: 6,
                  border: `1px solid ${item.isExtra ? 'var(--primary)' : '#E2E8F0'}`
                }}>
                  <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-secondary)' }}>{item.label}:</span>
                  <span style={{ fontSize: 12, fontWeight: 800, color: item.isExtra ? 'var(--primary)' : 'var(--text-primary)' }}>
                    {item.sems.length > 0 ? `S${item.sems.join(', S')}` : 'None'}
                  </span>
                </div>
              ));
            })()}
          </div>
        </div>
      )}

      {/* ── Metric Cards ── */}
      <div className="metrics-grid">
        <MetricCard label="Total Students" value={stats?.total_strength?.toLocaleString()} icon="🎓" color="#6366F1" />
        <MetricCard label="Unique Courses" value={stats?.unique_courses} icon="📚" color="#22D3EE" />
        <MetricCard label="Total Days" value={totalDays} icon="📅" color="#10B981" />
        <MetricCard label="Departments" value={stats?.dept_stats?.length ?? 0} icon="🏛️" color="#F59E0B" />
      </div>



      {/* ── Filter Bar ── */}
      <div className="filter-bar" style={{ 
        display: 'flex', 
        gap: 16, 
        marginBottom: 24, 
        padding: 20, 
        background: 'white', 
        borderRadius: 12, 
        border: '1px solid var(--border)',
        boxShadow: '0 2px 8px rgba(0,0,0,0.03)',
        alignItems: 'center',
        flexWrap: 'wrap'
      }}>
        <div style={{ flex: 1, minWidth: '250px', position: 'relative' }}>
          <span style={{ position: 'absolute', left: 14, top: '50%', transform: 'translateY(-50%)', opacity: 0.5 }}>🔍</span>
          <input 
            type="text" 
            placeholder="Search Display Day (e.g. '1')..." 
            className="form-input" 
            style={{ paddingLeft: 40, width: '100%', fontSize: 13 }}
            value={filterSearch}
            onChange={e => setFilterSearch(e.target.value)}
          />
        </div>
        
        <div style={{ display: 'flex', gap: 12 }}>
          <select 
            className="form-input" 
            style={{ width: '160px', fontSize: 13, cursor: 'pointer' }}
            value={filterDayType}
            onChange={e => setFilterDayType(e.target.value)}
          >
            <option value="all">📅 All Days</option>
            <option value="odd">🔢 Odd Days Only</option>
            <option value="even">🔢 Even Days Only</option>
          </select>

          <select 
            className="form-input" 
            style={{ width: '160px', fontSize: 13, cursor: 'pointer' }}
            value={filterSession}
            onChange={e => setFilterSession(e.target.value)}
          >
            <option value="all">🌓 All Sessions</option>
            <option value="FN">☀️ Forenoon (FN)</option>
            <option value="AN">🌙 Afternoon (AN)</option>
          </select>

          {(filterSearch || filterDayType !== 'all' || filterSession !== 'all') && (
            <button 
              onClick={() => { setFilterSearch(''); setFilterDayType('all'); setFilterSession('all'); }}
              style={{ background: 'none', border: 'none', color: '#6366F1', fontWeight: 700, fontSize: 12, cursor: 'pointer', padding: '0 8px' }}
            >
              Reset
            </button>
          )}
        </div>
      </div>

      {/* ── Generated Timetable Grid (DnD) ── */}
      <div className="tt-grid-section">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
          <h3 className="section-heading">🗓 Generated Schedule — Drag to Rearrange</h3>
          <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>Click a card to drill down · Drag to move between sessions</span>
        </div>

        {/* PDF-style table */}
        <div className="tt-pdf-table">
          <div className="tt-pdf-header">
            <div className="tt-pdf-col-left">Day — Session</div>
            <div className="tt-pdf-col-right">Course Code — Title</div>
          </div>

          <DndContext
            collisionDetection={closestCenter}
            onDragStart={e => setActiveId(e.active.id)}
            onDragOver={e => setOverSession(e.over?.id)}
            onDragEnd={handleDragEnd}
          >
            {[...days]
              .sort((a, b) => a - b)
              .filter(day => {
                if (filterSearch && !`Day ${day}`.toLowerCase().includes(filterSearch.toLowerCase()) && !String(day).includes(filterSearch)) return false;
                // Odd/Even physical filter
                if (filterDayType === 'odd' && day % 2 === 0) return false;
                if (filterDayType === 'even' && day % 2 !== 0) return false;
                return true;
              })
              .map(day => (
              ['FN', 'AN']
                .filter(sess => filterSession === 'all' || sess === filterSession)
                .map(session => {
                const sessionSlots = grouped[day]?.[session] || [];
                return (
                  <div key={`${day}-${session}`} className="tt-pdf-row-group">
                    <div className="tt-pdf-day-cell" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '150px' }}>
                      <span style={{ fontWeight: 900, fontSize: 20, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.5px' }}>Day {day}</span>
                      <span className="tt-session-pill"
                        style={{ background: session === 'FN' ? '#818CF8' : '#FB923C', margin: '8px 0' }}>
                        {session}
                      </span>
                      
                      {(() => {
                        // Priority 1: Use semesters from the actual courses in this session
                        const activeSems = new Set();
                        sessionSlots.forEach(s => {
                           (s.semesters || "").split(',').forEach(sem => { if(sem) activeSems.add(sem); });
                        });

                        if (activeSems.size > 0) {
                          return (
                            <div className="session-sem-badge" style={{ background: '#F1F5F9', color: '#475569', fontWeight: 800 }}>
                              <span>S{Array.from(activeSems).sort((a,b) => a-b).join(', ')}</span>
                            </div>
                          );
                        }

                        // Priority 2: Fallback to Day Configuration hint if empty
                        const dayType = (day % 2 !== 0) ? 'Odd' : 'Even';
                        const configKey = `${dayType}${session}`;
                        const configSems = timetable?.day_config?.[configKey] || [];
                        if (configSems.length === 0) return null;
                        return (
                          <div className="session-sem-badge">
                            <span>S{configSems.join(', ')}</span>
                          </div>
                        );
                      })()}

                      <button 
                         onClick={() => setSessionListView({ day, period: session })}
                         style={{ 
                            marginTop: 12,
                            padding: '6px 12px',
                            fontSize: 10,
                            fontWeight: 800,
                            background: 'white',
                            border: '1px solid #E2E8F0',
                            borderRadius: 6,
                            cursor: 'pointer',
                            color: '#64748B',
                            boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
                            transition: 'all 0.2s'
                         }}
                         onMouseOver={e => e.currentTarget.style.borderColor = '#94A3B8'}
                         onMouseOut={e => e.currentTarget.style.borderColor = '#E2E8F0'}
                      >
                         📋 DETAILS
                      </button>
                    </div>

                    <DroppableSession id={`${day}-${session}`}>
                      <SortableContext items={sessionSlots.map(s => s.id)} strategy={verticalListSortingStrategy}>
                        {sessionSlots.length === 0
                          ? <div className="tt-pdf-empty">No courses scheduled in this session</div>
                          : sessionSlots.map(slot => (
                            <CourseCard 
                              key={slot.id} 
                              slot={slot} 
                              onDrilldown={setDrilldown} 
                              isConflictState={getSlotConflict(slot.id)}
                              isMismatchState={getSlotMismatch(slot)}
                              initialSlots={initialSlots}
                            />
                          ))
                        }
                      </SortableContext>
                    </DroppableSession>
                  </div>
                );
              })
            ))}

            <DragOverlay>
              {activeId ? (
                <CourseCard
                  slot={activeSlot}
                  isConflictState={isConflict(overSession)}
                  isMismatchState={isMismatch(overSession)}
                  initialSlots={initialSlots}
                />
              ) : null}
            </DragOverlay>
          </DndContext>
        </div>
      </div>

      {/* ── Drilldown Modal ── */}
      {drilldown && (
        <div className="modal-overlay" onClick={() => setDrilldown(null)}>
          <div className="drilldown-modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 style={{ fontSize: 15 }}>📖 Course Details</h2>
              <button onClick={() => setDrilldown(null)} style={{ background: 'none', border: 'none', fontSize: 22, cursor: 'pointer', lineHeight: 1 }}>×</button>
            </div>
            <div style={{ padding: '16px 24px 24px' }}>
              <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 4 }}>{drilldown.course_name}</div>
              <div style={{ fontSize: 13, color: 'var(--primary)', marginBottom: 16, fontWeight: 600 }}>
                {drilldown.course_codes?.split(',').join(' / ')}
              </div>
              <div className="drilldown-meta">
                <div><span>📅 Slot</span> Day {drilldown.day_number} — {drilldown.session}</div>
                <div><span>👥 Strength</span> {drilldown.strength} students</div>
                <div><span>🏛️ Departments</span> {drilldown.departments}</div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Warning Clash Modal ── */}
      {warningModal && (
        <div className="modal-overlay" style={{ zIndex: 99999 }}>
          <div className="drilldown-modal" style={{ 
            borderTop: `4px solid ${warningModal.isConflict ? '#EF4444' : warningModal.isMismatch ? '#F59E0B' : '#6366F1'}`, 
            maxWidth: '500px' 
          }}>
            <div className="modal-header">
              <h2 style={{ 
                fontSize: 16, 
                color: warningModal.isConflict ? '#B91C1C' : warningModal.isMismatch ? '#92400E' : '#4338CA', 
                display: 'flex', 
                alignItems: 'center', 
                gap: 8 
              }}>
                {warningModal.isConflict ? '⚠️ Student Schedule Conflict' : 
                 warningModal.isMismatch ? 'ℹ️ Session Mismatch Warning' : 
                 'Confirm Manual Move'}
              </h2>
            </div>
            <div style={{ padding: '20px 24px 24px' }}>
              <div style={{ fontSize: 14, color: 'var(--text-primary)', lineHeight: 1.6 }}>
                {warningModal.isConflict && warningModal.isMismatch ? (
                  <p>This move is <strong>highly discouraged</strong>. It causes a student schedule overlap AND places the exam in a non-optimal session.</p>
                ) : warningModal.clashingStudents?.length > 0 ? (
                  <div style={{ color: '#B91C1C' }}>
                    <p style={{ fontWeight: 700, marginBottom: 8 }}>⚠️ CRITICAL CONFLICT</p>
                    <p>Moving this course will cause <strong>{warningModal.clashingStudents.length} students</strong> to have two exams scheduled at the same time:</p>
                    <div style={{ 
                      maxHeight: '120px', 
                      overflowY: 'auto', 
                      background: '#FEF2F2', 
                      padding: '8px 12px', 
                      borderRadius: 6, 
                      marginTop: 8,
                      fontSize: 12,
                      border: '1px solid #FCA5A5'
                    }}>
                      {warningModal.clashingStudents.map((name, i) => (
                        <div key={i} style={{ padding: '2px 0' }}>• {name}</div>
                      ))}
                    </div>
                  </div>
                ) : warningModal.isConflict ? (
                  <p><strong>Warning:</strong> Students from shared departments are already writing exams in this session. Proceeding will cause a conflict.</p>
                ) : warningModal.isMismatch ? (
                  <p><strong>Note:</strong> This session type (Day {warningModal.toSlot.day_number} {warningModal.toSlot.session}) is not configured for {warningModal.fromSlot.course_name}&apos;s semester.</p>
                ) : (
                  <p>Are you sure you want to move <strong>{warningModal.fromSlot.course_name}</strong> to this session?</p>
                )}
                
                <div style={{ background: '#F8FAFC', padding: 12, borderRadius: 8, marginTop: 16, fontSize: 13, border: '1px solid var(--border)' }}>
                   Target: <strong>Day {warningModal.toSlot.day_number} — {warningModal.toSlot.session}</strong>
                </div>
              </div>

              <div style={{ display: 'flex', gap: 12, marginTop: 24 }}>
                <button 
                  className="btn-primary" 
                  style={{ 
                    background: warningModal.isConflict ? '#EF4444' : warningModal.isMismatch ? '#F59E0B' : '#6366F1', 
                    borderColor: 'transparent', 
                    flex: 1 
                  }} 
                  onClick={warningModal.onConfirm}
                >
                  {warningModal.isConflict || warningModal.isMismatch ? 'Move Anyway' : 'Confirm Move'}
                </button>
                <button className="btn-secondary" style={{ flex: 1 }} onClick={() => setWarningModal(null)}>Cancel</button>
              </div>
            </div>
          </div>
        </div>
      )}
      {/* ── Session Student List Modal ── */}
      {sessionListView && (
        <SessionDetailsModal
          timetableId={timetableId || timetable?.id}
          day={sessionListView.day}
          period={sessionListView.period}
          onClose={() => setSessionListView(null)}
        />
      )}
    </main>
  );
}
