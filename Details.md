# Bit-Timetable: Professional Exam Scheduling Engine 📅🚀

This document serves as the comprehensive "Source of Truth" for the **Bit-Timetable** project. It is designed to provide full context, architectural deep-dives, and feature mapping for any AI agent or developer stepping into the codebase.

---

## 1. What & Why? (The Purpose)
**Bit-Timetable** is a high-performance, intelligent scheduling application designed for academic institutions to generate conflict-free exam timetables. 

### The Problem it Solves:
- **Manual Effort**: Traditional scheduling takes days and is prone to human error (student clashes).
- **Complex Constraints**: Balancing Regular vs. Arrear exams, Honors/Minors priority, and department-level distribution.
- **Data Silos**: Previously relied on external drives; now uses a secure, local-first architecture.

---

## 2. The Tech Stack
- **Frontend**: React (Vite) with a premium, custom CSS design system (Vanilla CSS).
- **Backend**: Go (Gin Gonic) for a lightning-fast concurrent scheduling engine.
- **Database**: MySQL (XAMPP) for persistent storage of student data and generated timetables.
- **Utilities**: `excelie` (Go) for parsing, `lucide-react` for icons, `dnd-kit` for interactive dragging.

---

## 3. The Whole Story (Evolution)
1. **The Inception**: Started as a tool to map students to slots for a fixed regulation.
2. **Local Revolution**: Migrated from Google Drive to local file parsing (Excel/CSV) for speed and privacy.
3. **The Multi-Mode Engine**: Refactored the core logic into a **4-Mode Selector** (1, 2, 4, or 5 semesters).
4. **Smart Filler (Mode 5)**: Developed a "Gap-Filler" logic where a 5th semester (Extra) fills gaps in the primary schedule and takes over after the main phase ends.
5. **Quality of Life**: Added a high-fidelity Dashboard with Drag-and-Drop editing and "Mismatch Glow" visual feedback.

---

## 4. How it Works: The Engine (`scheduler.go`)
The engine uses a **Multi-Pass Greedy Algorithm with Pre-Sorting**.

### Phase A: Pre-Processing
- **Clubbing**: Courses with the same name across different departments/sections are merged if they don't have clashing students, creating a unified "Session Slot".
- **Sorting (Regular)**: `Honors` ➔ `Minors` ➔ `Leader Dept Extra Electives (Solo)` ➔ `Leader Dept Extra Core (Suffix Order like 706)` ➔ `Standard Balancing`.
- **Sorting (Arrear)**: Strength-based (`High to Low` or `Low to High`) - Bypasses all leader/priority logic.

### Phase B: Generation passes
- **Pass 1 (Primary)**: Schedules semesters (S1-S4) into FN/AN sessions.
- **Pass 2 (Extra/Overflow)**: 
  - **Gap Filling**: If a session in Phase 1 is empty, the 5th semester fills it (if no clashing students).
  - **Density**: 5th semester courses now "Club" together multiple subjects in one session just like primary semesters.
  - **Standalone Electives**: If the 5th semester is filling gaps as a "Leader", its Electives may be scheduled Alone first.
  - **Post-Main Takeover**: Once Phase 1 semesters (1-4) finish all their exams, the 5th semester automatically takes over **both FN and AN sessions every day** until exhausted.

---

## 5. UI Feature Map (Button-by-Button)

### A. Data Import Page
- **"Choose File"**: Parses local `.xlsx` or `.csv` using `excelize` logic.
- **"Import Data"**: Cleans course codes (alphanumeric only), parses Roman numeral semesters, and saves to MySQL.
- **Smart Scrubbing**: The backend automatically cleans `-`, `*`, and spaces from course codes on startup but only for "dirty" rows (optimized for speed).

### B. Create Timetable Page
- **Mode Selector**: 
  - `1 Sem`: Daily FN+AN scheduling (same semester both sessions).
  - `2 Sems`: Daily Split (Semester A in FN, Semester B in AN).
  - `4 Sems`: Alternating Odd/Even logic (OddFN/AN, EvenFN/AN).
  - `5 Sems`: 4 Sems + Extra "Smart Filler" semester.
- **Design**: All cards are left-aligned with consistent professional styling (no dashed borders).
- **"+ Generate"**: Triggers the Go engine and opens the Draft Dashboard.

### C. Dashboard (Draft Review)
- **Output-Centric Metrics**: "Unique Courses" and "Total Students" stats are calculated purely from the **generated timetable ID**, not the global database.
- **Unique Departments**: Fixed logic to split comma-separated strings (e.g. "CSE,IT") to count individual departments correctly (e.g. 30 instead of 97).
- **Configuration Reference Bar**: A visual summary (e.g., "Odd FN: S6") shown above metrics for easy verification.
- **Visual Cues (Drag & Drop UX)**: 
  - `Red Background (Conflict)`: Student Clash (Two courses have overlapping register numbers in the same session).
  - `Yellow Glow (Mismatch)`: Course moved to a session it wasn't originally intended for (e.g. S4 in an Odd day).
  - `Zebra Striped (Conflict + Mismatch)`: A high-alert state where a course is both in the wrong session AND clashing with students (Red/Yellow repeating diagonal gradient).
  - `Green Highlight (Manual Move)`: Shows that a course has been manually moved from its generated position to a **safe, conflict-free** new session.
- **Semester Badges**: Every card has a small badge (e.g. `S5`) in the corner for instant identification.

### D. Time Tables Page (The Archives)
- **List View**: Displays all approved timetables with **Semester Badges** (extracts `S1, S3, S5` from JSON config).
- **"📝 Edit" (Revert)**: Moves a saved timetable back into "Pending" status for live modification.
- **"🗑 Delete"**: Permanent removal from the database.

---

## 6. Key Logic Rules (For Reference)
- **Honors/Minors**: In Regular exams, these always take absolute priority to ensure they finish early.
- **Leader Logic**: If Dept A has 7 exams and Dept B has 6, Dept A is a "Leader". Its 7th exam (if Elective) will be scheduled **Alone** in a session to balance the load.
- **Mode 5 Post-Main**: Once semesters 1-4 finish their final exam of the set, the 5th semester schedules in **both FN and AN** every day until it finishes.

---

## 7. Developer Notes
- **API Base**: `http://localhost:8080/api`.
- **CORS**: Enabled for `http://localhost:5173`.
- **Critical Tables**: `student_data` (Input), `timetables` (Metadata), `timetable_slots` (Result Output).

---
*Created by Antigravity AI - Powering Intelligent Academic Logistics.*
