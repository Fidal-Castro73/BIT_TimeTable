import React, { useState } from 'react';
import './index.css';
import Dashboard from './Dashboard';

// --- Icons ---
const BookIcon = ({ size = 24 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
);

const CalendarIcon = ({ size = 20 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
);

const GearIcon = ({ size = 24 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
);

const PlusIcon = ({ size = 16 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
);

const ArrowLeftIcon = ({ size = 16 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="19" y1="12" x2="5" y2="12"></line><polyline points="12 19 5 12 12 5"></polyline></svg>
);

const ChevronDownIcon = ({ size = 16 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>
);

const SearchIcon = ({ size = 16 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
);

const FilterIcon = ({ size = 20 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon></svg>
);

const ImportIcon = ({ size = 20 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
);

const DriveIcon = ({ size = 20 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M7 2a2 2 0 0 1 2 2v1h6V4a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2V5H9v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h2z"></path><path d="M12 11V7"></path><path d="M12 11a5 5 0 0 1 5 5v3a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-3a5 5 0 0 1 5-5z"></path></svg>
);

const FileIcon = ({ size = 20 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
);

const TablePlusIcon = ({ size = 64 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="3" width="18" height="15" rx="2" ry="2" />
    <line x1="3" y1="8" x2="21" y2="8" />
    <line x1="3" y1="13" x2="21" y2="13" />
    <line x1="9" y1="3" x2="9" y2="18" />
    <line x1="15" y1="3" x2="15" y2="18" />
    <circle cx="18" cy="18" r="5" fill="var(--bg-card)" stroke="none" />
    <circle cx="18" cy="18" r="5" />
    <line x1="18" y1="15.5" x2="18" y2="20.5" />
    <line x1="15.5" y1="18" x2="20.5" y2="18" />
  </svg>
);

// --- Components ---

function Sidebar({ activeView, onViewChange }) {
  const [isExpanded, setIsExpanded] = useState(true);

  return (
    <aside className="sidebar">
      <div className="sidebar-header" style={{ cursor: 'pointer' }} onClick={() => onViewChange('welcome')}>
        <div className="brand-icon">
          <BookIcon size={20} />
        </div>
        <div className="brand-text-container">
          <span className="brand-title">ACADEMICS</span>
          <span className="brand-subtitle">Portal</span>
        </div>
      </div>

      <div className="sidebar-section">
        <div className="section-label">MENU</div>

        <div className="menu-group">
          <div className="menu-header" onClick={() => setIsExpanded(!isExpanded)}>
            <div className="menu-header-content">
              <span className="menu-item-icon">
                <CalendarIcon size={18} />
              </span>
              <span>Time Table</span>
            </div>
            <div style={{ transform: isExpanded ? 'rotate(180deg)' : 'none', transition: 'transform 0.2s' }}>
              <ChevronDownIcon size={14} />
            </div>
          </div>

          {isExpanded && (
            <div className="submenu">
              <div
                className={`submenu-item ${activeView === 'list' || activeView === 'create' ? 'active' : ''}`}
                onClick={() => onViewChange('list')}
              >
                Time Tables
              </div>
              <div
                className={`submenu-item ${activeView === 'dashboard' ? 'active' : ''}`}
                onClick={() => onViewChange('dashboard')}
              >
                Dashboard
              </div>
              <div
                className={`submenu-item ${activeView === 'import' ? 'active' : ''}`}
                onClick={() => onViewChange('import')}
              >
                Import Data
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="sidebar-footer">
        <div className="user-avatar">A</div>
        <div className="user-info">
          <span className="user-name">admin</span>
          <span className="user-email">admin@example.com</span>
        </div>
      </div>
    </aside>
  );
}

function WelcomeScreen() {
  return (
    <main className="main-content">
      <div className="welcome-screen">
        <div className="welcome-icon">
          <BookIcon size={40} />
        </div>
        <h1>Welcome Admin</h1>
        <p>Manage your academic timetables, exam configurations, and student data all in one place. Select a menu item to get started.</p>
      </div>
    </main>
  );
}

function ListTimetables({ onViewChange, onSelectTimetable }) {
  const [timetables, setTimetables] = React.useState([]);
  const [search, setSearch] = React.useState('');

  React.useEffect(() => {
    fetch('http://localhost:8080/api/timetables?status=all')
      .then(r => r.json())
      .then(d => d.success && setTimetables(d.timetables || []))
      .catch(() => { });
  }, []);

  const handleDelete = async (id, e) => {
    e.stopPropagation();
    if (!window.confirm('Delete this timetable from the database?')) return;
    await fetch(`http://localhost:8080/api/timetable/${id}`, { method: 'DELETE' });
    setTimetables(prev => prev.filter(t => t.id !== id));
  };

  const handleRevert = async (id, e) => {
    e.stopPropagation();
    try {
      const res = await fetch(`http://localhost:8080/api/timetable/${id}/revert`, { method: 'PUT' });
      const data = await res.json();
      if (data.success) {
        onViewChange('dashboard', id);
      } else {
        alert("Failed to revert: " + data.error);
      }
    } catch (err) {
      alert("Network error. Please try again.");
    }
  };

  const filtered = timetables.filter(t =>
    t.name?.toLowerCase().includes(search.toLowerCase()) ||
    String(t.regulation_id).includes(search)
  );

  return (
    <main className="main-content">
      <div style={{ marginBottom: 32 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 20, color: 'var(--text-primary)' }}>Timetable Management</h1>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div className="search-input-wrapper" style={{ position: 'relative', width: 320 }}>
            <div style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-secondary)' }}>
              <SearchIcon size={18} />
            </div>
            <input type="text" className="form-input search-input" placeholder="Search by name or regulation..."
              style={{ paddingLeft: 44, width: '100%', borderRadius: 20, background: '#F1F5F9', border: '1px solid transparent' }}
              value={search} onChange={e => setSearch(e.target.value)}
            />
          </div>
          <button className="btn-primary" onClick={() => onViewChange('create')}>+ Create Timetable</button>
        </div>
      </div>

      {filtered.length === 0 ? (
        <div className="card">
          <div className="card-body" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '64px 24px', minHeight: 300 }}>
            <div style={{ color: '#94A3B8', marginBottom: 20 }}><TablePlusIcon size={72} /></div>
            <p style={{ color: 'var(--text-secondary)', marginBottom: 24, fontSize: 16, fontWeight: 500 }}>No timetables found. Generate one first!</p>
            <button className="btn-outline" onClick={() => onViewChange('create')}>Create Timetable</button>
          </div>
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
          {filtered.map(tt => (
            <div key={tt.id} className="tt-list-card" onClick={() => { onSelectTimetable(tt); onViewChange('view-timetable'); }}>
              <div className="tt-list-card-icon">{tt.status === 'approved' ? '📅' : '⏳'}</div>
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <div className="tt-list-card-name">{tt.name}</div>
                  {tt.status === 'pending' && (
                    <span style={{ fontSize: 10, background: '#FEF3C7', color: '#92400E', padding: '2px 8px', borderRadius: 10, fontWeight: 700 }}>PENDING REVIEW</span>
                  )}
                </div>
                <div className="tt-list-card-meta">Regulation {tt.regulation_id}</div>
                
                {/* Semester Extraction logic */}
                {(() => {
                  try {
                    const cfg = JSON.parse(tt.day_config || '{}');
                    const sems = new Set();
                    Object.values(cfg).forEach(v => {
                      if (Array.isArray(v)) v.forEach(s => sems.add(s));
                      if (v.FN && Array.isArray(v.FN)) v.forEach(s => sems.add(s)); // Safety for depth
                      if (v.AN && Array.isArray(v.AN)) v.forEach(s => sems.add(s));
                      // Handle nested mode objects
                      if (v.FN) {
                         if(Array.isArray(v.FN)) v.FN.forEach(s => sems.add(s));
                      }
                      if (v.AN) {
                         if(Array.isArray(v.AN)) v.AN.forEach(s => sems.add(s));
                      }
                    });
                    // Fallback for ExtraSem
                    if(cfg.extra_sem) sems.add(cfg.extra_sem);
                    if(cfg.extraSem) sems.add(cfg.extraSem);

                    const sortedSems = Array.from(sems).filter(Boolean).sort((a,b) => a-b);
                    if (sortedSems.length > 0) {
                      return (
                        <div style={{ marginTop: 4, display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                          <span style={{ fontSize: 10, color: 'var(--text-secondary)', fontWeight: 600 }}>Semesters:</span>
                          {sortedSems.map(s => (
                            <span key={s} style={{ fontSize: 10, background: '#E2E8F0', color: '#475569', padding: '1px 6px', borderRadius: 4, fontWeight: 700 }}>S{s}</span>
                          ))}
                        </div>
                      )
                    }
                  } catch(e) {}
                  return null;
                })()}

                <div className="tt-list-card-date">Generated on {new Date(tt.generated_at).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })}</div>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8, alignItems: 'flex-end' }}>
                <div style={{ display: 'flex', gap: 12 }}>
                  <button
                    className="tt-edit-btn"
                    style={{ background: 'none', border: 'none', color: 'var(--primary)', cursor: 'pointer', fontSize: 13, fontWeight: 700 }}
                    onClick={e => handleRevert(tt.id, e)}
                  >
                    📝 Edit
                  </button>
                  <button className="tt-delete-btn" onClick={e => handleDelete(tt.id, e)}>🗑 Delete</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </main>
  );
}

function ViewTimetable({ timetable, onBack }) {
  const [slots, setSlots] = React.useState([]);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    if (!timetable?.id) return;
    setLoading(true);
    fetch(`http://localhost:8080/api/timetable/${timetable.id}`)
      .then(r => r.json())
      .then(d => { if (d.success) setSlots(d.slots || []); })
      .finally(() => setLoading(false));
  }, [timetable?.id]);

  // Group slots by day + session
  const grouped = {};
  for (const slot of slots) {
    const key = slot.day_number;
    if (!grouped[key]) grouped[key] = { FN: [], AN: [] };
    grouped[key][slot.session].push(slot);
  }
  const days = Object.keys(grouped).map(Number).sort((a, b) => a - b);

  return (
    <main className="main-content">
      <div className="page-header">
        <a href="#" className="back-link" onClick={e => { e.preventDefault(); onBack(); }}>
          <ArrowLeftIcon /> Back to Time Tables
        </a>
        <h1 style={{ marginTop: 12 }}>{timetable?.name}</h1>
        <p>Regulation {timetable?.regulation_id} · Generated {new Date(timetable?.generated_at).toLocaleDateString()}</p>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 60, color: 'var(--text-secondary)' }}>Loading timetable...</div>
      ) : (
        <div style={{ background: 'white', padding: 32, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <span style={{ fontWeight: 700, fontSize: 18, color: 'black' }}>BIT Timetable</span>
            <a
              href={`http://localhost:8080/api/timetable/${timetable?.id}/pdf`}
              target="_blank"
              rel="noopener noreferrer"
              className="btn-primary"
              style={{ textDecoration: 'none', fontSize: 13, padding: '8px 18px', background: 'black', color: 'white', border: 'none', borderRadius: 4 }}
            >
              📄 Download PDF
            </a>
          </div>

          <table style={{ width: '100%', borderCollapse: 'collapse', border: '2px solid black', fontFamily: 'Arial, sans-serif', color: 'black' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid black' }}>
                <th style={{ borderRight: '2px solid black', padding: '10px 16px', textAlign: 'center', fontWeight: 'bold', fontSize: 14, width: '25%' }}>Exam. Date -<br />Session</th>
                <th style={{ padding: '10px 16px', textAlign: 'center', fontWeight: 'bold', fontSize: 14 }}>Course Code - Title</th>
              </tr>
            </thead>
            <tbody>
              {days.map(day => (
                ['FN', 'AN'].map((session) => {
                  const sessionSlots = grouped[day]?.[session] || [];
                  if (sessionSlots.length === 0) return null; // exact matching image means skipping empty
                  return sessionSlots.map((slot, idx) => (
                    <tr key={slot.id}>
                      {idx === 0 && (
                        <td rowSpan={sessionSlots.length} style={{ borderBottom: '2px solid black', borderRight: '2px solid black', padding: '12px 16px', textAlign: 'center', fontWeight: 'bold', fontSize: 13, verticalAlign: 'middle' }}>
                          Day {day} - {session}
                        </td>
                      )}
                      <td style={{ borderBottom: (idx === sessionSlots.length - 1) ? '2px solid black' : '1px solid black', padding: '10px 12px', fontSize: 13 }}>
                        <span style={{ fontWeight: 400 }}>
                          {slot.course_codes?.split(',').join(' / ')} - {slot.course_name?.toUpperCase()}
                        </span>
                      </td>
                    </tr>
                  ));
                })
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
  );
}

import * as XLSX from 'xlsx';

function ImportDataView() {
  const [regulation, setRegulation] = useState('');
  const [activeCategory, setActiveCategory] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [uploadTab, setUploadTab] = useState('local');
  const [imports, setImports] = useState([]); // List of { regulation, category, fileName, recordsCount, status: 'ready' | 'processing' | 'imported' }
  const [editingIndex, setEditingIndex] = useState(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const fileInputRef = React.useRef(null);

  React.useEffect(() => {
    const fetchHistory = async () => {
      try {
        const res = await fetch("http://localhost:8080/api/imports");
        const data = await res.json();
        if (data.success) {
          setImports(data.imports || []);
        }
      } catch (err) {
        console.error("Failed to load import history:", err);
      }
    };
    fetchHistory();
  }, []);

  const categories = ['Regular', 'Arrear'];

  const handleBrowseClick = () => {
    fileInputRef.current?.click();
  };

  const finalizeImport = (fileObj, fileName = 'uploaded_file.xlsx', recordsCount = 0) => {
    const newEntry = {
      regulation,
      category: activeCategory,
      fileName,
      recordsCount,
      status: 'ready',
      fileObj: fileObj
    };

    if (editingIndex !== null) {
      const updatedImports = [...imports];
      updatedImports[editingIndex] = newEntry;
      setImports(updatedImports);
      setEditingIndex(null);
    } else {
      const exists = imports.find(item => item.regulation.toString() === regulation.toString() && item.category === activeCategory);
      if (exists) {
        alert(`Records for ${activeCategory} Course Regulation ${regulation} are already uploaded. Please remove the existing entry from the summary below before uploading a new file.`);
        return;
      }
      setImports([...imports, newEntry]);
    }

    setIsModalOpen(false);
    setRegulation('');
    setActiveCategory(null);
  };

  const handleFileUpload = (e) => {
    if (!regulation || !activeCategory) return;
    const file = e.target.files?.[0];
    if (!file) return;

    // Use XLSX to parse the local file for record count perfectly
    const reader = new FileReader();
    reader.onload = (evt) => {
      try {
        const bstr = evt.target.result;
        const wb = XLSX.read(bstr, { type: 'binary' });
        const wsname = wb.SheetNames[0];
        const ws = wb.Sheets[wsname];
        const data = XLSX.utils.sheet_to_json(ws);
        finalizeImport(file, file.name, data.length);
      } catch (err) {
        console.error("Parse Error:", err);
        finalizeImport(file, file.name); // Fallback to random count
      }
    };
    reader.readAsBinaryString(file);
  };



  const handleProcessAll = async () => {
    if (imports.filter(m => m.status === 'ready' || m.status === 'error').length === 0) return;
    setIsProcessing(true);

    for (let i = 0; i < imports.length; i++) {
      const item = imports[i];
      if (item.status === 'ready' || item.status === 'error') {

        setImports(prev => {
          const next = [...prev];
          next[i] = { ...next[i], status: 'processing' };
          return next;
        });

        try {
          const formData = new FormData();
          formData.append('file', item.fileObj);
          formData.append('regulation_id', item.regulation);
          formData.append('upload_type', item.category);

          const res = await fetch("http://localhost:8080/api/upload", {
            method: "POST",
            body: formData
          });

          const data = await res.json();

          if (data.success) {
            setImports(prev => {
              const next = [...prev];
              next[i] = { ...next[i], status: 'imported', recordsCount: data.inserted };
              return next;
            });
          } else {
            alert("Error on record " + item.regulation + ": " + data.error);
            setImports(prev => {
              const next = [...prev];
              next[i] = { ...next[i], status: 'error' };
              return next;
            });
          }
        } catch (error) {
          alert("Network link error on " + item.regulation + ". Make sure Go backend is running on 8080 !");
          setImports(prev => {
            const next = [...prev];
            next[i] = { ...next[i], status: 'error' };
            return next;
          });
        }
      }
    }

    setIsProcessing(false);
  };

  const handleEdit = (index) => {
    const item = imports[index];
    setRegulation(item.regulation);
    setActiveCategory(item.category);
    setEditingIndex(index);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const cancelEdit = () => {
    setEditingIndex(null);
    setRegulation('');
    setActiveCategory(null);
  };

  const handleSaveInfo = async () => {
    if (!regulation || !activeCategory) return;
    const item = imports[editingIndex];

    try {
      const res = await fetch("http://localhost:8080/api/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          old_regulation: parseInt(item.regulation),
          old_category: item.category,
          new_regulation: parseInt(regulation),
          new_category: activeCategory
        })
      });

      const data = await res.json();
      if (data.success) {
        const updatedImports = [...imports];
        updatedImports[editingIndex] = { ...updatedImports[editingIndex], regulation, category: activeCategory };
        setImports(updatedImports);
        setEditingIndex(null);
        setRegulation('');
        setActiveCategory(null);
      } else {
        alert("Update failed: " + data.error);
      }
    } catch (err) {
      alert("Sync error: Unable to update records in database.");
    }
  };

  const removeImport = async (index) => {
    const item = imports[index];

    if (window.confirm(`Are you sure you want to remove ${item.category} records for Regulation ${item.regulation}? This will delete all associated data from the database.`)) {
      try {
        const res = await fetch(`http://localhost:8080/api/delete?regulation_id=${item.regulation}&upload_type=${item.category}`, {
          method: "DELETE"
        });

        const data = await res.json();
        if (data.success) {
          if (editingIndex === index) {
            cancelEdit();
          }
          setImports(imports.filter((_, i) => i !== index));
        } else {
          alert("Delete failed: " + data.error);
        }
      } catch (err) {
        alert("Sync error: Unable to delete records from database.");
      }
    }
  };



  const UploadModal = () => (
    <div className="modal-overlay" onClick={() => setIsModalOpen(false)}>
      <div className="upload-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 style={{ fontSize: 16, fontWeight: 700 }}>
            {editingIndex !== null ? 'Replace File' : 'Select File to Import'}
          </h2>
          <button
            onClick={() => setIsModalOpen(false)}
            style={{ background: 'none', border: 'none', fontSize: 20, cursor: 'pointer', color: 'var(--text-secondary)' }}
          >
            &times;
          </button>
        </div>

        <div className="card-body" style={{ padding: '0 24px 24px' }}>
          <div className="dropzone" style={{ padding: '40px 24px', margin: '24px 0 0' }}>
            <div className="dropzone-icon" style={{ width: 48, height: 48 }}>
              <FileIcon size={24} />
            </div>
            <div style={{ textAlign: 'center' }}>
              <p style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: 14 }}>
                Drag & drop files here
              </p>
              <p style={{ fontSize: 12 }}>or click to browse from computer</p>
            </div>
            <input
              type="file"
              ref={fileInputRef}
              style={{ display: 'none' }}
              accept=".xlsx,.xls,.csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-excel"
              onChange={handleFileUpload}
            />
            <button className="btn-primary" style={{ marginTop: 8, fontSize: 13 }} onClick={handleBrowseClick}>
              {editingIndex !== null ? 'Replace File' : 'Browse Files'}
            </button>
          </div>

          <div style={{ marginTop: 24 }}>
            <h3 style={{ fontSize: 12, fontWeight: 700, marginBottom: 12, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
              📋 Required Excel Format (Sample)
            </h3>
            <div style={{ overflowX: 'auto', border: '1px solid var(--border)', borderRadius: 8 }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
                <thead style={{ background: '#F8FAFC' }}>
                  <tr>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', color: 'var(--text-secondary)' }}>A</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.1)', color: 'var(--primary)', fontWeight: 700 }}>B</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.1)', color: 'var(--primary)', fontWeight: 700 }}>C</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.1)', color: 'var(--primary)', fontWeight: 700 }}>D</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.1)', color: 'var(--primary)', fontWeight: 700 }}>E</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.1)', color: 'var(--primary)', fontWeight: 700 }}>F</th>
                  </tr>
                  <tr>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left' }}>S.NO</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.05)' }}>COURSE_CODE</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.05)' }}>COURSE_NAME</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.05)' }}>SEMESTER</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.05)' }}>REGISTER_NO</th>
                    <th style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', textAlign: 'left', background: 'rgba(99, 102, 241, 0.05)' }}>STUDENT_NAME</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)' }}>1</td>
                    <td style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', background: 'rgba(99, 102, 241, 0.02)' }}>18CS401</td>
                    <td style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', background: 'rgba(99, 102, 241, 0.02)' }}>MATHS IV</td>
                    <td style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', background: 'rgba(99, 102, 241, 0.02)' }}>4</td>
                    <td style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', background: 'rgba(99, 102, 241, 0.02)' }}>7376...</td>
                    <td style={{ padding: '8px 12px', borderBottom: '1px solid var(--border)', background: 'rgba(99, 102, 241, 0.02)' }}>JOHN DOE</td>
                  </tr>
                </tbody>

              </table>
            </div>
            <p style={{ marginTop: 12, fontSize: 11, lineHeight: '1.5', color: 'var(--text-secondary)' }}>
              <span style={{ color: 'var(--primary)', fontWeight: 700 }}>Note:</span> Ensure your data starts from the second row (Row 2). The first row (Row 1) should be headers.
            </p>
          </div>


          <div style={{ padding: '24px 0 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <p style={{ fontSize: 11, color: 'var(--text-secondary)' }}>
              Supported: .xlsx, .xls, .csv
            </p>
            <button className="btn-secondary" style={{ fontSize: 13 }} onClick={() => setIsModalOpen(false)}>Cancel</button>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <main className="main-content">
      <div className="page-header">
        <h1>Import Data</h1>
        <p>Sophisticated data ingestion for academic records</p>
      </div>

      <div className={`card ${editingIndex !== null ? 'editing-mode' : ''}`} style={editingIndex !== null ? { borderColor: 'var(--primary)', boxShadow: '0 0 0 1px var(--primary)' } : {}}>
        <div className="card-header">
          <div className="card-icon-container" style={{ width: 44, height: 44, background: editingIndex !== null ? 'var(--primary-light)' : '' }}>
            <ImportIcon size={22} color={editingIndex !== null ? 'var(--primary)' : 'currentColor'} />
          </div>
          <div className="card-title-group">
            <h2 style={{ fontSize: 16 }}>{editingIndex !== null ? 'Edit Import Configuration' : 'Data Import Configuration'}</h2>
            <p style={{ fontSize: 13 }}>{editingIndex !== null ? 'Modify existing criteria or replace the file.' : 'Define criteria and upload excel/sheet files.'}</p>
          </div>
          {editingIndex !== null && (
            <button
              className="btn-secondary"
              style={{ marginLeft: 'auto', padding: '6px 12px', fontSize: 12 }}
              onClick={cancelEdit}
            >
              Cancel Edit
            </button>
          )}
        </div>
        <div className="card-body">
          <div className="form-group">
            <label className="form-label">Regulation Year <span className="required">*</span></label>
            <input
              type="text"
              className="form-input"
              placeholder="e.g. 2021"
              value={regulation}
              onChange={(e) => setRegulation(e.target.value)}
              style={{ maxWidth: '400px' }}
            />
          </div>

          {regulation && (
            <div style={{ marginTop: 24 }}>
              <div className="form-group" style={{ maxWidth: '400px' }}>
                <label className="form-label">Select Course Category</label>
                <div className="select-wrapper">
                  <select
                    className="form-select"
                    value={activeCategory || ""}
                    onChange={(e) => setActiveCategory(e.target.value)}
                  >
                    <option value="" disabled>Select Course Category</option>
                    {categories.map(cat => (
                      <option key={cat} value={cat}>{cat} Course</option>
                    ))}
                  </select>
                  <div className="select-icon">
                    <ChevronDownIcon size={14} />
                  </div>
                </div>
              </div>

              {activeCategory && (
                <div className="inline-dropzone-container">
                  <div className="inline-dropzone-label">
                    {editingIndex !== null ? 'Modify Information or Replace File' : 'Upload Excel file'} <span className="required">*</span>
                  </div>
                  <div style={{ display: 'flex', gap: '12px', alignItems: 'stretch', maxWidth: '600px' }}>
                    <div className="inline-dropzone" onClick={() => setIsModalOpen(true)} style={{ flex: 1, margin: 0 }}>
                      <button className="choose-file-btn" onClick={(e) => { e.stopPropagation(); setIsModalOpen(true); }}>
                        {editingIndex !== null ? 'Change File' : 'Choose File'}
                      </button>
                      <div className="inline-dropzone-text">
                        {editingIndex !== null
                          ? `Replace file for ${activeCategory} (${regulation})`
                          : `Upload records for ${activeCategory} Course (${regulation})`}
                      </div>
                    </div>
                    {editingIndex !== null && (
                      <button
                        className="btn-primary"
                        style={{ padding: '0 20px', fontSize: 13, height: '48px' }}
                        onClick={handleSaveInfo}
                      >
                        Save Changes
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {imports.length > 0 && (
        <div style={{ marginTop: 32 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--text-primary)' }}>Summary of Scheduled Imports</h3>
            <button
              className="btn-primary"
              style={{ padding: '8px 20px', fontSize: 13 }}
              onClick={handleProcessAll}
              disabled={isProcessing || imports.every(i => i.status === 'imported')}
            >
              {isProcessing ? 'Processing...' : 'Process All Imports'}
            </button>
          </div>
          <div className="card" style={{ maxWidth: 'none' }}>
            <div className="card-body" style={{ padding: 0 }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead style={{ background: '#F8FAFC' }}>
                  <tr>
                    <th style={{ textAlign: 'left', padding: '12px 24px', fontSize: 11, color: 'var(--text-secondary)', borderBottom: '1px solid var(--border)', textTransform: 'uppercase' }}>Regulation</th>
                    <th style={{ textAlign: 'left', padding: '12px 24px', fontSize: 11, color: 'var(--text-secondary)', borderBottom: '1px solid var(--border)', textTransform: 'uppercase' }}>Category</th>
                    <th style={{ textAlign: 'left', padding: '12px 24px', fontSize: 11, color: 'var(--text-secondary)', borderBottom: '1px solid var(--border)', textTransform: 'uppercase' }}>File Name</th>
                    <th style={{ textAlign: 'center', padding: '12px 24px', fontSize: 11, color: 'var(--text-secondary)', borderBottom: '1px solid var(--border)', textTransform: 'uppercase' }}>Records</th>
                    <th style={{ textAlign: 'center', padding: '12px 24px', fontSize: 11, color: 'var(--text-secondary)', borderBottom: '1px solid var(--border)', textTransform: 'uppercase' }}>Status</th>
                    <th style={{ textAlign: 'right', padding: '12px 24px', fontSize: 11, color: 'var(--text-secondary)', borderBottom: '1px solid var(--border)', textTransform: 'uppercase' }}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {imports.map((item, index) => (
                    <tr key={index} style={editingIndex === index ? { background: 'var(--primary-light)' } : {}}>
                      <td style={{ padding: '16px 24px', fontSize: 13, borderBottom: '1px solid var(--border)', fontWeight: editingIndex === index ? 600 : 400 }}>{item.regulation}</td>
                      <td style={{ padding: '16px 24px', fontSize: 14, borderBottom: '1px solid var(--border)', fontWeight: editingIndex === index ? 600 : 400 }}>{item.category} Course</td>
                      <td style={{ padding: '16px 24px', fontSize: 13, borderBottom: '1px solid var(--border)', color: 'var(--text-secondary)' }}>{item.fileName}</td>
                      <td style={{ padding: '16px 24px', fontSize: 13, borderBottom: '1px solid var(--border)', textAlign: 'center' }}>
                        <span style={{ padding: '2px 8px', background: '#F1F5F9', borderRadius: '12px', fontWeight: 600 }}>{item.recordsCount}</span>
                      </td>
                      <td style={{ padding: '16px 24px', textAlign: 'center', borderBottom: '1px solid var(--border)' }}>
                        <span style={{
                          padding: '4px 10px',
                          borderRadius: '20px',
                          fontSize: 11,
                          fontWeight: 700,
                          textTransform: 'uppercase',
                          background: item.status === 'imported' ? '#DCFCE7' : item.status === 'processing' ? '#FEF9C3' : '#F1F5F9',
                          color: item.status === 'imported' ? '#166534' : item.status === 'processing' ? '#854D0E' : '#475569'
                        }}>
                          {item.status}
                        </span>
                      </td>
                      <td style={{ padding: '16px 24px', textAlign: 'right', borderBottom: '1px solid var(--border)' }}>
                        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
                          <button
                            style={{ color: 'var(--primary)', background: 'none', border: 'none', cursor: 'pointer', fontSize: 12, fontWeight: 500 }}
                            onClick={() => handleEdit(index)}
                          >
                            Edit
                          </button>
                          <button
                            style={{ color: '#EF4444', background: 'none', border: 'none', cursor: 'pointer', fontSize: 12, fontWeight: 500 }}
                            onClick={() => removeImport(index)}
                          >
                            Remove
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {isModalOpen && <UploadModal />}
    </main>
  );
}


function CreateTimetable({ onViewChange }) {
  const [examName, setExamName] = React.useState('');
  const [regulation, setRegulation] = React.useState('');
  const [generating, setGenerating] = React.useState(false);
  const [pendingDraft, setPendingDraft] = React.useState(null);
  const [semCount, setSemCount] = React.useState(4);
  const [selections, setSelections] = React.useState({
    day1FN: { semester: '', type: 'Regular' },
    day1AN: { semester: '', type: 'Regular' },
    day2FN: { semester: '', type: 'Regular' },
    day2AN: { semester: '', type: 'Regular' },
    single: { semester: '', type: 'Regular' },
    splitFN: { semester: '', type: 'Regular' },
    splitAN: { semester: '', type: 'Regular' },
    extra: { semester: '', type: 'Regular' },
  });

  const [showSettings, setShowSettings] = React.useState(false);
  const [genSettings, setGenSettings] = React.useState({
    regular: 'asc', // Low to High
    arrear: 'desc'  // High to Low
  });

  React.useEffect(() => {
    fetch('http://localhost:8080/api/timetables?status=pending')
      .then(r => r.json())
      .then(data => {
        if (data.success && data.timetables?.length > 0) {
          setPendingDraft(data.timetables[0]);
        }
      });
  }, []);

  const semesters = [1, 2, 3, 4, 5, 6, 7, 8];
  const types = ['Regular', 'Arrear'];

  const handleSelect = (key, field, value) => {
    setSelections(prev => ({ ...prev, [key]: { ...prev[key], [field]: value } }));
  };

  const SemesterField = ({ label, selectionKey }) => (
    <div className="form-group" style={{ marginBottom: 16 }}>
      <label className="form-label">{label}</label>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
        <div className="select-wrapper">
          <select
            className="form-select"
            value={selections[selectionKey].semester}
            onChange={(e) => handleSelect(selectionKey, 'semester', e.target.value)}
          >
            <option value="">Select Semester</option>
            {semesters.map(s => (
              <option key={s} value={s}>Semester {s}</option>
            ))}
          </select>
          <div className="select-icon"><ChevronDownIcon /></div>
        </div>
        <div className="select-wrapper">
          <select
            className="form-select"
            value={selections[selectionKey].type}
            onChange={(e) => handleSelect(selectionKey, 'type', e.target.value)}
          >
            {types.map(t => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
          <div className="select-icon"><ChevronDownIcon /></div>
        </div>
      </div>
    </div>
  );

  const parseSems = (key) => {
    const v = selections[key].semester;
    return v ? [parseInt(v)] : [];
  };

  const handleGenerate = async () => {
    if (!examName || !regulation) {
      alert('Please fill in Exam Name and Regulation Year.');
      return;
    }
    // Transform selections based on mode
    let dayConfig = {};
    if (semCount === 1) {
      const s = parseSems('single');
      const t = selections.single.type;
      dayConfig = { OddFN: s, OddAN: s, EvenFN: s, EvenAN: s, odd_fn_type: t, odd_an_type: t, even_fn_type: t, even_an_type: t, mode: 1 };
    } else if (semCount === 2) {
      const sFN = parseSems('splitFN');
      const sAN = parseSems('splitAN');
      const tFN = selections.splitFN.type;
      const tAN = selections.splitAN.type;
      dayConfig = { OddFN: sFN, OddAN: sAN, EvenFN: sFN, EvenAN: sAN, odd_fn_type: tFN, odd_an_type: tAN, even_fn_type: tFN, even_an_type: tAN, mode: 2 };
    } else if (semCount === 4) {
      dayConfig = {
        OddFN: parseSems('day1FN'), OddAN: parseSems('day1AN'),
        EvenFN: parseSems('day2FN'), EvenAN: parseSems('day2AN'),
        odd_fn_type: selections.day1FN.type, odd_an_type: selections.day1AN.type,
        even_fn_type: selections.day2FN.type, even_an_type: selections.day2AN.type,
        mode: 4
      };
    } else if (semCount === 5) {
      dayConfig = {
        OddFN: parseSems('day1FN'), OddAN: parseSems('day1AN'),
        EvenFN: parseSems('day2FN'), EvenAN: parseSems('day2AN'),
        odd_fn_type: selections.day1FN.type, odd_an_type: selections.day1AN.type,
        even_fn_type: selections.day2FN.type, even_an_type: selections.day2AN.type,
        extra_sem: parseSems('extra')[0] || 0,
        extra_type: selections.extra.type,
        mode: 5
      };
    }

    setGenerating(true);
    try {
      const res = await fetch('http://localhost:8080/api/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: examName,
          regulation_id: parseInt(regulation),
          day_config: dayConfig,
          sort_options: genSettings,
        }),
      });
      const data = await res.json();
      if (data.success) {
        alert(`✅ ${data.message}`);
        onViewChange('dashboard', data.timetable_id);
      } else {
        alert('Generation failed: ' + data.error);
      }
    } catch (err) {
      alert('Network error. Make sure Go backend is running!');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <main className="main-content">
      <div className="page-header">
        <h1>Timetable Management</h1>
        <p>Create and manage academic exam timetables</p>
        <a href="#" className="back-link" onClick={(e) => { e.preventDefault(); onViewChange('list'); }}>
          <ArrowLeftIcon /> Back to Time Tables
        </a>
      </div>

      {pendingDraft && (
        <div className="card" style={{ borderColor: '#F59E0B', background: '#FFFBEB', marginBottom: 24, padding: 20 }}>
          <div style={{ display: 'flex', gap: 16 }}>
            <div style={{ fontSize: 24 }}>⚠️</div>
            <div>
              <h3 style={{ color: '#92400E', fontSize: 14, fontWeight: 700 }}>Action Required: Pending Draft Found</h3>
              <p style={{ color: '#B45309', fontSize: 13, marginTop: 4 }}>
                You already have a pending timetable (<strong>{pendingDraft.name}</strong>) in the dashboard.
                Please Approve or Discard it before creating a new one.
              </p>
              <button
                className="btn-primary"
                style={{ marginTop: 12, background: '#D97706', padding: '6px 14px', fontSize: 12 }}
                onClick={() => onViewChange('dashboard', pendingDraft.id)}
              >
                Go to Dashboard
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="card">
        <div className="card-header">
          <div className="card-icon-container"><BookIcon size={24} /></div>
          <div className="card-title-group">
            <h2>New Exam Details</h2>
            <p>Define the general details for this timetable.</p>
          </div>
        </div>
        <div className="card-body">
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
            <div className="form-group" style={{ marginBottom: 0 }}>
              <label className="form-label">Exam Name <span className="required">*</span></label>
              <input type="text" className="form-input" placeholder="e.g. NOV 2025 Exam" style={{ maxWidth: 'none' }} value={examName} onChange={e => setExamName(e.target.value)} />
            </div>
            <div className="form-group" style={{ marginBottom: 0 }}>
              <label className="form-label">Regulation <span className="required">*</span></label>
              <input type="text" className="form-input" placeholder="e.g. 2021" style={{ maxWidth: 'none' }} value={regulation} onChange={e => setRegulation(e.target.value)} />
            </div>
          </div>
        </div>
      </div>

      <div className="section-title" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <div className="section-title-icon"><GearIcon size={22} /></div>
          <h3>Schedule Configuration</h3>
        </div>
        <div className="mode-selector" style={{ display: 'flex', gap: 8, background: '#F1F5F9', padding: 4, borderRadius: 8 }}>
          {[1, 2, 4, 5].map(m => (
            <button
              key={m}
              onClick={() => setSemCount(m)}
              style={{
                padding: '6px 16px',
                borderRadius: 6,
                border: 'none',
                fontSize: 12,
                fontWeight: 700,
                cursor: 'pointer',
                background: semCount === m ? 'white' : 'transparent',
                color: semCount === m ? 'var(--primary)' : 'var(--text-secondary)',
                boxShadow: semCount === m ? '0 2px 4px rgba(0,0,0,0.05)' : 'none'
              }}
            >
              {m === 1 ? '1 Sem' : m === 2 ? '2 Sems' : m === 4 ? '4 Sems' : '5 Sems'}
            </button>
          ))}
        </div>
      </div>

      <div className="schedule-grid" style={{
        display: 'grid',
        gridTemplateColumns: (semCount === 1 || semCount === 2) ? '1fr' : '1fr 1fr',
        gap: 20
      }}>
        {semCount === 1 && (
          <div className="day-card" style={{ maxWidth: '400px' }}>
            <div className="day-header">DAILY (Every Session)</div>
            <div className="day-body">
              <SemesterField label="Target Semester" selectionKey="single" />
            </div>
          </div>
        )}

        {semCount === 2 && (
          <div className="day-card" style={{ maxWidth: '600px' }}>
            <div className="day-header">DAILY (Split Sessions)</div>
            <div className="day-body" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
              <SemesterField label="Forenoon (FN)" selectionKey="splitFN" />
              <SemesterField label="Afternoon (AN)" selectionKey="splitAN" />
            </div>
          </div>
        )}

        {(semCount === 4 || semCount === 5) && (
          <>
            <div className="day-card">
              <div className="day-header">ODD DAYS</div>
              <div className="day-body">
                <SemesterField label="Forenoon (FN)" selectionKey="day1FN" />
                <SemesterField label="Afternoon (AN)" selectionKey="day1AN" />
              </div>
            </div>
            <div className="day-card">
              <div className="day-header">EVEN DAYS</div>
              <div className="day-body">
                <SemesterField label="Forenoon (FN)" selectionKey="day2FN" />
                <SemesterField label="Afternoon (AN)" selectionKey="day2AN" />
              </div>
            </div>
          </>
        )}
      </div>

      {semCount === 5 && (
        <div className="day-card" style={{ marginTop: 20 }}>
          <div className="day-header">ADDITIONAL SEMESTER</div>
          <div className="day-body" style={{ maxWidth: '400px' }}>
            <SemesterField label="Extra Semester" selectionKey="extra" />
          </div>
        </div>
      )}

      <div className="page-footer">
        <button
          className="btn-secondary"
          style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#64748B' }}
          onClick={() => setShowSettings(true)}
        >
          <GearIcon size={16} /> Strategy Settings
        </button>
        <div style={{ flex: 1 }}></div>
        <button className="btn-secondary" onClick={() => onViewChange('list')}>Cancel</button>
        <button className="btn-primary" onClick={handleGenerate} disabled={generating || pendingDraft}>
          <CalendarIcon /> {generating ? 'Generating...' : 'Generate Timetable'}
        </button>
      </div>

      {showSettings && (
        <div className="modal-overlay" onClick={() => setShowSettings(false)}>
          <div className="upload-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: 450 }}>
            <div className="modal-header">
              <h2 style={{ fontSize: 16, fontWeight: 700 }}>⚙️ Generation Strategy</h2>
              <button onClick={() => setShowSettings(false)} style={{ background: 'none', border: 'none', fontSize: 20, cursor: 'pointer' }}>&times;</button>
            </div>
            <div style={{ padding: 24 }}>
              <div style={{ marginBottom: 24 }}>
                <label className="form-label" style={{ display: 'block', marginBottom: 12 }}>Regular Exam Priority</label>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                  <button
                    className={genSettings.regular === 'asc' ? 'btn-primary' : 'btn-secondary'}
                    style={{ fontSize: 12, padding: '10px' }}
                    onClick={() => setGenSettings(p => ({ ...p, regular: 'asc' }))}
                  >
                    Low ➜ High Strength
                  </button>
                  <button
                    className={genSettings.regular === 'desc' ? 'btn-primary' : 'btn-secondary'}
                    style={{ fontSize: 12, padding: '10px' }}
                    onClick={() => setGenSettings(p => ({ ...p, regular: 'desc' }))}
                  >
                    High ➜ Low Strength
                  </button>
                </div>
              </div>

              <div style={{ marginBottom: 12 }}>
                <label className="form-label" style={{ display: 'block', marginBottom: 12 }}>Arrear Exam Priority</label>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                  <button
                    className={genSettings.arrear === 'asc' ? 'btn-primary' : 'btn-secondary'}
                    style={{ fontSize: 12, padding: '10px' }}
                    onClick={() => setGenSettings(p => ({ ...p, arrear: 'asc' }))}
                  >
                    Low ➜ High Strength
                  </button>
                  <button
                    className={genSettings.arrear === 'desc' ? 'btn-primary' : 'btn-secondary'}
                    style={{ fontSize: 12, padding: '10px' }}
                    onClick={() => setGenSettings(p => ({ ...p, arrear: 'desc' }))}
                  >
                    High ➜ Low Strength
                  </button>
                </div>
              </div>

              <p style={{ marginTop: 24, fontSize: 11, color: '#94A3B8', fontStyle: 'italic', textAlign: 'center' }}>
                Note: These settings determine the scheduling order during the generation process.
              </p>
            </div>
          </div>
        </div>
      )}
    </main>

  );
}

import { ErrorBoundary } from './ErrorBoundary';

function App() {
  const [view, setView] = useState('welcome');
  const [selectedTimetable, setSelectedTimetable] = useState(null);
  const [pendingTimetableId, setPendingTimetableId] = useState(null);

  const handleViewChange = (newView) => setView(newView);

  return (
    <div className="app-container">
      <Sidebar activeView={view} onViewChange={handleViewChange} />
      {view === 'welcome' && <WelcomeScreen />}
      {view === 'list' && <ListTimetables onViewChange={handleViewChange} onSelectTimetable={setSelectedTimetable} />}
      {view === 'view-timetable' && selectedTimetable && <ViewTimetable timetable={selectedTimetable} onBack={() => setView('list')} />}
      {view === 'dashboard' && (
        <ErrorBoundary>
          <Dashboard
            timetableId={pendingTimetableId}
            onApproved={() => { setPendingTimetableId(null); handleViewChange('list'); }}
            onRejected={() => setPendingTimetableId(null)}
            onViewChange={handleViewChange}
          />
        </ErrorBoundary>
      )}
      {view === 'create' && <CreateTimetable onViewChange={(v, id) => {
        if (v === 'dashboard' && id) {
          setPendingTimetableId(id);
        }
        handleViewChange(v);
      }} />}
      {view === 'import' && <ImportDataView />}
    </div>
  );
}

export default App;
