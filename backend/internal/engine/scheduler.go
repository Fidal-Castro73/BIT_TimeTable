package engine

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"bittimetable/internal/clash"
)

// CourseInfo holds all metadata for a single course
type CourseInfo struct {
	CourseCode   string
	CourseName   string
	CourseType   string // Core, Elective, Honors, Minors, Open Elective
	Department   string
	Semester     int   // primary semester (for sorting)
	Semesters    []int // all semesters this course spans across departments
	Strength     int
	RegulationID int
	UploadType   string
	StudentSet   map[string]bool // THE ROSTER (Who writes the exam in this session)
	ClashSet     map[string]bool // THE SHADOW (Everyone in the course, used for safety)

	// Advanced tracking for dynamic roster construction
	CodeToGlobal map[string]map[string]bool
	CodeToArrear map[string]map[string]bool
	CodeToSem    map[string]int
	CodeToDept   map[string]string
}

// SlotCourse represents a single (possibly clubbed) course entry in a slot
type SlotCourse struct {
	CourseCodes []string // multiple if clubbed
	CourseName  string
	Strength    int
	Departments []string
	Semesters   []int
	StudentSets []map[string]bool // one per code (The Roster)
	ClashSets   []map[string]bool // one per code (The Safety Set)
	Type        string            // "Regular" or "Arrear"
}

// ScheduleSlot is one session in the final timetable
type ScheduleSlot struct {
	DayNumber int
	Session   string // FN or AN
	Courses   []SlotCourse
}

// DayConfig maps semester numbers to their assigned session per day type
// e.g. OddFN=[1,3], OddAN=[5,7], EvenFN=[2,4], EvenAN=[6,8]
type DayConfig struct {
	OddFN      []int
	OddAN      []int
	EvenFN     []int
	EvenAN     []int
	OddFNType  string `json:"odd_fn_type"`
	OddANType  string `json:"odd_an_type"`
	EvenFNType string `json:"even_fn_type"`
	EvenANType string `json:"even_an_type"`
	Mode       int    `json:"mode"`
	ExtraSem   int    `json:"extra_sem"`
	ExtraType  string `json:"extra_type"`
}

// SortOptions defines the strength sorting preference for each category
// "asc" for Low-to-High, "desc" for High-to-Low
type SortOptions struct {
	Regular string `json:"regular"`
	Arrear  string `json:"arrear"`
}

// loadCourses fetches all distinct courses for a regulation+uploadType with their metadata
func loadCourses(db *sql.DB, regulationID int, uploadType string) ([]CourseInfo, error) {
	rows, err := db.Query(`
		SELECT 
			course_code, 
			MAX(course_name), 
			MAX(course_type), 
			GROUP_CONCAT(DISTINCT department SEPARATOR ', '), 
			GROUP_CONCAT(DISTINCT semester SEPARATOR ','),
			COUNT(*) as strength
		FROM student_data
		WHERE regulation_id = ? AND upload_type = ?
		GROUP BY course_code
	`, regulationID, uploadType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []CourseInfo
	for rows.Next() {
		var c CourseInfo
		var courseType sql.NullString
		var semStr string
		if err := rows.Scan(&c.CourseCode, &c.CourseName, &courseType, &c.Department, &semStr, &c.Strength); err != nil {
			continue
		}
		if courseType.Valid {
			c.CourseType = courseType.String
		} else {
			c.CourseType = "Core"
		}
		c.RegulationID = regulationID
		c.UploadType = uploadType

		// Parse all semesters from GROUP_CONCAT
		c.Semesters = []int{}
		for _, s := range strings.Split(semStr, ",") {
			s = strings.TrimSpace(s)
			if v, err := strconv.Atoi(s); err == nil {
				c.Semesters = append(c.Semesters, v)
			}
		}
		if len(c.Semesters) > 0 {
			c.Semester = c.Semesters[0]
		}

		// Skip Non-Examination courses (Laboratory or Project)
		nameClean := strings.ToLower(strings.TrimSpace(c.CourseName))
		if strings.HasSuffix(nameClean, "laboratory") || strings.Contains(nameClean, "project") {
			continue
		}

		// Setup advanced map tracking
		cleanCode := CleanCourseCode(c.CourseCode)
		// fullSet = EVERYONE for this code regardless of upload type/semester
		// Used for CodeToGlobal so Primary Sem includes both Regular and Arrear students.
		// The Semester Firewall in dynRoster building then blocks wrong-semester Regulars.
		fullSet, _ := clash.GetStudentSet(db, c.CourseCode, regulationID, false)
		// arrSet = only Arrear students for this code (for Arrear-visitor enrollment)
		arrSet, _ := clash.GetStudentSet(db, c.CourseCode, regulationID, true)
		// regSet = only same-pool students, for correct initial strength count
		regSet, _ := clash.GetStudentSetForUploadType(db, c.CourseCode, regulationID, uploadType, c.Semester)

		c.CodeToGlobal = make(map[string]map[string]bool)
		c.CodeToArrear = make(map[string]map[string]bool)
		c.CodeToSem = make(map[string]int)
		c.CodeToDept = make(map[string]string)

		// CodeToGlobal uses fullSet so that in a Regular session, S4 Arrears
		// can still join S4 Regulars. The Semester Firewall (in Generate)
		// handles blocking S6 Regular students from appearing.
		c.CodeToGlobal[cleanCode] = fullSet
		c.CodeToArrear[cleanCode] = arrSet
		c.CodeToSem[cleanCode] = c.Semester
		c.CodeToDept[cleanCode] = c.Department

		c.ClashSet = fullSet
		// StudentSet: use regSet for accurate strength; Arrear pool uses arrSet
		if uploadType == "Arrear" {
			c.StudentSet = arrSet
		} else {
			c.StudentSet = regSet
		}
		c.Strength = len(c.StudentSet)
		courses = append(courses, c)
	}

	// ── Merge courses with the same name into one entry ──────────────────
	// e.g. "Digital Signal Processing" with codes 22CS601, 22EC602, 22IT603
	// should be scheduled as a single exam with all codes/departments combined.
	courses = mergeByName(courses)

	return courses, nil
}

// mergeByName collapses all CourseInfo entries that share the same course name
// (case-insensitive) into a single entry, merging codes, depts, semesters and student sets.
func mergeByName(courses []CourseInfo) []CourseInfo {
	type mergedEntry struct {
		idx  int
		seen map[string]bool // dept dedup
		sems map[int]bool    // sem dedup
	}
	nameIndex := map[string]int{} // normalised name → index in merged
	var merged []CourseInfo

	for _, c := range courses {
		// Aggressive Normalization: Remove all non-alphanumeric chars and internal spaces for comparison
		reg := strings.NewReplacer(" ", "", "-", "", "/", "", "(", "", ")", "", "_", "")
		key := strings.ToLower(reg.Replace(c.CourseName))

		if idx, exists := nameIndex[key]; exists {
			// Merge into existing entry
			m := &merged[idx]

			// Add course code if not already present (Aggressive Normalization)
			cleanNewCode := CleanCourseCode(c.CourseCode)

			alreadyCode := false
			for _, code := range strings.Split(m.CourseCode, ",") {
				if CleanCourseCode(code) == cleanNewCode {
					alreadyCode = true
					break
				}
			}
			if !alreadyCode && cleanNewCode != "" {
				m.CourseCode = m.CourseCode + "," + cleanNewCode
			}

			// Merge departments (dedup via the string itself)
			existingDepts := strings.Split(m.Department, ", ")
			deptSet := make(map[string]bool)
			for _, d := range existingDepts {
				deptSet[strings.TrimSpace(d)] = true
			}
			for _, d := range strings.Split(c.Department, ", ") {
				d = strings.TrimSpace(d)
				if d != "" && !deptSet[d] {
					m.Department = m.Department + ", " + d
					deptSet[d] = true
				}
			}

			// Merge semesters (dedup)
			semSet := make(map[int]bool)
			for _, s := range m.Semesters {
				semSet[s] = true
			}
			for _, s := range c.Semesters {
				if !semSet[s] {
					m.Semesters = append(m.Semesters, s)
					semSet[s] = true
				}
			}

			// Merge student sets (union — for clash detection)
			if m.StudentSet == nil {
				m.StudentSet = make(map[string]bool)
			}
			for reg := range c.StudentSet {
				m.StudentSet[reg] = true
			}
			if m.ClashSet == nil {
				m.ClashSet = make(map[string]bool)
			}
			for reg := range c.ClashSet {
				m.ClashSet[reg] = true
			}

			// Merge advanced maps
			if m.CodeToGlobal == nil { m.CodeToGlobal = make(map[string]map[string]bool) }
			if m.CodeToArrear == nil { m.CodeToArrear = make(map[string]map[string]bool) }
			if m.CodeToSem == nil { m.CodeToSem = make(map[string]int) }
			if m.CodeToDept == nil { m.CodeToDept = make(map[string]string) }

			for code, rmap := range c.CodeToGlobal { m.CodeToGlobal[code] = rmap }
			for code, rmap := range c.CodeToArrear { m.CodeToArrear[code] = rmap }
			for code, sem := range c.CodeToSem { m.CodeToSem[code] = sem }
			for code, dept := range c.CodeToDept { m.CodeToDept[code] = dept }

			// Use unique headcount for merged strength
			m.Strength = len(m.StudentSet)
		} else {
			// New entry
			nameIndex[key] = len(merged)
			merged = append(merged, c)
		}
	}

	return merged
}

// CleanCourseCode strictly removes EVERYTHING except letters and numbers
func CleanCourseCode(code string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return reg.ReplaceAllString(code, "")
}

// extractCodeSuffix — gets the numeric suffix from a course code for Core sequencing
// e.g. "18BT301" → 301, "22CSI601" → 601
func extractCodeSuffix(code string) int {
	// Find the last sequence of digits
	suffix := ""
	for i := len(code) - 1; i >= 0; i-- {
		if code[i] >= '0' && code[i] <= '9' {
			suffix = string(code[i]) + suffix
		} else {
			if suffix != "" {
				break
			}
		}
	}
	val, _ := strconv.Atoi(suffix)
	return val
}

// sortCoursesRegular — applies Regular exam priority ordering
func sortCoursesRegular(courses []CourseInfo, order string) []CourseInfo {
	if len(courses) == 0 {
		return courses
	}

	// Group by department → count courses
	deptCourses := make(map[string][]int) // indices
	deptFullCount := make(map[string]int)
	for i, c := range courses {
		deptCourses[c.Department] = append(deptCourses[c.Department], i)
		deptFullCount[c.Department]++
	}

	// Determine the "Baseline" count (what most depts have)
	// We'll use the minimum count as the baseline for "Extra" detection
	minCount := 999
	for _, cnt := range deptFullCount {
		if cnt < minCount {
			minCount = cnt
		}
	}

	// Identify "Extra" courses for each dept
	isExtra := make([]bool, len(courses))
	for _, indices := range deptCourses {
		if len(indices) > minCount {
			// Sort the department's indices by Type (Elective first) then by Suffix (Highest first for Core)
			sort.Slice(indices, func(i, j int) bool {
				a, b := courses[indices[i]], courses[indices[j]]
				if a.CourseType != b.CourseType {
					if a.CourseType == "Elective" { return true }
					if b.CourseType == "Elective" { return false }
				}
				// If both core, highest suffix first (e.g. 706 before 701)
				return extractCodeSuffix(a.CourseCode) > extractCodeSuffix(b.CourseCode)
			})
			
			// Mark the first N - minCount as "Extra"
			for i := 0; i < (len(indices) - minCount); i++ {
				isExtra[indices[i]] = true
			}
		}
	}

	sort.SliceStable(courses, func(i, j int) bool {
		a, b := courses[i], courses[j]
		aExtra, bExtra := isExtra[i], isExtra[j]

		// Priority 1: Honors first
		aIsHonors := a.CourseType == "Honors"
		bIsHonors := b.CourseType == "Honors"
		if aIsHonors != bIsHonors {
			return aIsHonors
		}

		// Priority 2: Minors second
		aIsMinors := a.CourseType == "Minors"
		bIsMinors := b.CourseType == "Minors"
		if aIsMinors != bIsMinors {
			return aIsMinors
		}

		// Priority 3: Leader's Extra Electives Absolute Priority
		aIsExtraElective := aExtra && a.CourseType == "Elective"
		bIsExtraElective := bExtra && b.CourseType == "Elective"
		if aIsExtraElective != bIsExtraElective {
			return aIsExtraElective
		}

		// Priority 4: Leader's Extra Core Absolute Priority (706 logic)
		aIsExtraCore := aExtra && a.CourseType == "Core"
		bIsExtraCore := bExtra && b.CourseType == "Core"
		if aIsExtraCore != bIsExtraCore {
			return aIsExtraCore
		}

		// Priority 5: Sort by department size (bigger dept first for balancing)
		if deptFullCount[a.Department] != deptFullCount[b.Department] {
			return deptFullCount[a.Department] > deptFullCount[b.Department]
		}

		// Priority 6: Within same dept — Strength based
		if a.Strength != b.Strength {
			if order == "desc" {
				return a.Strength > b.Strength
			}
			return a.Strength < b.Strength
		}

		// Priority 7: Core courses — follow suffix order (descending for Core tails, ascending for rest)
		// We already handled extra core above. Rest follow ascending.
		return extractCodeSuffix(a.CourseCode) < extractCodeSuffix(b.CourseCode)
	})

	// Add a temporary flag/metadata if needed, but we'll use a hack in Generate:
	// Courses with Priority level 'Extra Elective' should break the clubbing loop.
	return courses
}

// sortCoursesArrear — applies Arrear priority ordering
func sortCoursesArrear(courses []CourseInfo, order string) []CourseInfo {
	sort.SliceStable(courses, func(i, j int) bool {
		if order == "asc" {
			return courses[i].Strength < courses[j].Strength
		}
		return courses[i].Strength > courses[j].Strength
	})
	return courses
}

// canClubWith checks if a candidate course can share a session with all already-placed courses
func canClubWith(candidate CourseInfo, slotCourses []SlotCourse) bool {
	for _, sc := range slotCourses {
		for i, existingSet := range sc.ClashSets {
			if clash.HasClash(candidate.ClashSet, existingSet) {
				intersection := clash.Intersection(candidate.ClashSet, existingSet)
				fmt.Printf("🛑 CLASH DETECTED: %s vs %s | Student: %s (and %d others)\n", 
					candidate.CourseCode, sc.CourseCodes[i], intersection[0], len(intersection)-1)
				return false
			}
		}
	}
	return true
}

// buildSessions expands day config into ordered []{DayNumber, Session} pairs
func buildSessions(config DayConfig, totalDaysNeeded int) []struct {
	Day     int
	Session string
} {
	var sessions []struct {
		Day     int
		Session string
	}

	day := 1
	for len(sessions) < totalDaysNeeded*2+10 { // +buffer
		isOdd := (day % 2) != 0

		if isOdd {
			if len(config.OddFN) > 0 {
				sessions = append(sessions, struct {
					Day     int
					Session string
				}{day, "FN"})
			}
			sessions = append(sessions, struct {
				Day     int
				Session string
			}{day, "AN"})
		} else {
			if len(config.EvenFN) > 0 {
				sessions = append(sessions, struct {
					Day     int
					Session string
				}{day, "FN"})
			}
			sessions = append(sessions, struct {
				Day     int
				Session string
			}{day, "AN"})
		}
		day++
	}
	return sessions
}

// Generate runs the full scheduling algorithm and returns a list of ScheduleSlots
func Generate(db *sql.DB, regulationID int, config DayConfig, options SortOptions) ([]ScheduleSlot, error) {
	// 1. Load BOTH Regular and Arrear courses if needed
	regCourses, _ := loadCourses(db, regulationID, "Regular")
	arrCourses, _ := loadCourses(db, regulationID, "Arrear")

	if len(regCourses) == 0 && len(arrCourses) == 0 {
		return nil, fmt.Errorf("no courses found for regulation %d", regulationID)
	}

	// 2. Sort both pools according to user preferences
	regCourses = sortCoursesRegular(regCourses, options.Regular)
	arrCourses = sortCoursesArrear(arrCourses, options.Arrear)

	// SEMESTER FIREWALL: find each Regular student's NATIVE semester
	// (the semester where they have the most course enrollments).
	// Used to definitively block e.g. S6 Regular students from appearing
	// in S4 sessions, regardless of per-code logic.
	// We store: roll -> native_semester (ONE value per student).
	studentNativeSem := make(map[string]int)
	{
		rows, err := db.Query(`
			SELECT register_no, semester, COUNT(*) as cnt
			FROM student_data
			WHERE regulation_id = ? AND upload_type = 'Regular'
			GROUP BY register_no, semester
			ORDER BY register_no, cnt DESC`, regulationID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var roll string
				var sem, cnt int
				if rows.Scan(&roll, &sem, &cnt) == nil {
					// Only store the first (highest-count) semester per student
					if _, exists := studentNativeSem[roll]; !exists {
						studentNativeSem[roll] = sem
					}
				}
			}
		}
	}

	// Track which courses are scheduled (Map by Course Code)
	regScheduled := make(map[string]bool)
	arrScheduled := make(map[string]bool)
	// perCourseScheduled: courseName -> set of rolls already placed for that specific course.
	// This prevents the same student from writing the same exam TWICE (e.g., once in an Arrear
	// session and again in a Regular session for the same-named course).
	// It does NOT block students from their other courses across other days.
	perCourseScheduled := make(map[string]map[string]bool)
	totalScheduled := 0
	totalToSchedule := len(regCourses) + len(arrCourses)

	var result []ScheduleSlot
	// Calculate buffer days based on total exams
	sessions := buildSessions(config, totalToSchedule)

	for _, sess := range sessions {
		if totalScheduled >= totalToSchedule {
			break
		}

		// Determine which semesters AND Category (Reg/Arr) are valid for this session
		var validSemesters []int
		var targetType string
		isOdd := (sess.Day % 2) != 0
		if isOdd {
			if sess.Session == "FN" {
				validSemesters = config.OddFN
				targetType = config.OddFNType
			} else {
				validSemesters = config.OddAN
				targetType = config.OddANType
			}
		} else {
			if sess.Session == "FN" {
				validSemesters = config.EvenFN
				targetType = config.EvenFNType
			} else {
				validSemesters = config.EvenAN
				targetType = config.EvenANType
			}
		}

		// Fallback for older configs
		if targetType == "" {
			targetType = "Regular"
		}

		// If this session has no semester assignments, skip
		if len(validSemesters) == 0 {
			continue
		}

		validSemSet := make(map[int]bool)
		for _, s := range validSemesters {
			validSemSet[s] = true
		}

		slot := ScheduleSlot{DayNumber: sess.Day, Session: sess.Session}

		// Use the correct pool based on targetType
		var pool []CourseInfo
		if targetType == "Regular" {
			pool = regCourses
		} else {
			pool = arrCourses
		}

		// --- Pass 1: Primary Semesters (S1-S4) ---
		for i := 0; i < len(pool); i++ {
			course := pool[i]
			courseKey := strings.ToUpper(strings.TrimSpace(course.CourseName))
			alreadyInCourse := perCourseScheduled[courseKey] // nil map is safe for reads

			if isCourseScheduled(course, targetType, regScheduled, arrScheduled) {
				continue
			}

			// Check semester eligibility
			semMatch := false
			for _, sem := range course.Semesters {
				if validSemSet[sem] {
					semMatch = true
					break
				}
			}
			if !semMatch {
				continue
			}

			// -- DYNAMIC ROSTER RECONSTRUCTION --
			dynRoster := make(map[string]bool)
			presentCodesMap := make(map[string]bool)

			for code, baseSem := range course.CodeToSem {
				codeContributed := false
				if targetType == "Regular" {
					if validSemSet[baseSem] { // Primary Sem
						if !regScheduled[code] {
							for s := range course.CodeToGlobal[code] {
								if alreadyInCourse[s] { continue } // already wrote THIS course
								if nativeSem, isRegular := studentNativeSem[s]; isRegular {
									if !validSemSet[nativeSem] { continue }
								}
								dynRoster[s] = true
								codeContributed = true
							}
						}
					} else { // Secondary Sem
						if !arrScheduled[code] {
							for s := range course.CodeToArrear[code] {
								if !alreadyInCourse[s] { 
									dynRoster[s] = true 
									codeContributed = true
								}
							}
						}
					}
				} else { // Arrear session
					if !arrScheduled[code] {
						for s := range course.CodeToArrear[code] {
							if !alreadyInCourse[s] { 
								dynRoster[s] = true 
								codeContributed = true
							}
						}
					}
				}
				if codeContributed {
					presentCodesMap[code] = true
				}
			}

			// Construct the list of codes, semesters, and depts actually present in this session
			var activeCodes []string
			activeSemMap := make(map[int]bool)
			activeDeptMap := make(map[string]bool)

			for code := range presentCodesMap {
				activeCodes = append(activeCodes, code)
				if sem, ok := course.CodeToSem[code]; ok {
					activeSemMap[sem] = true
				}
				if dept, ok := course.CodeToDept[code]; ok {
					// Handle comma separated departments
					for _, d := range strings.Split(dept, ",") {
						activeDeptMap[strings.TrimSpace(d)] = true
					}
				}
			}
			
			var activeSemesters []int
			for sem := range activeSemMap {
				activeSemesters = append(activeSemesters, sem)
			}
			var activeDepartments []string
			for dept := range activeDeptMap {
				activeDepartments = append(activeDepartments, dept)
			}

			// Fallback: if somehow empty but we allow placement
			if len(activeCodes) == 0 && len(dynRoster) > 0 {
				activeCodes = strings.Split(course.CourseCode, "+")
				activeSemesters = course.Semesters
				activeDepartments = strings.Split(course.Department, ", ")
			}

			if len(dynRoster) == 0 {
				continue
			}

			course.StudentSet = dynRoster
			course.Strength = len(dynRoster)

			if !canClubWith(course, slot.Courses) {
				continue
			}

			// Clubbing logic...
			clubbed := false
			for sci := range slot.Courses {
				if strings.EqualFold(slot.Courses[sci].CourseName, course.CourseName) {
					// Add only active codes and semesters to prevent showing codes/semesters with 0 students
					slot.Courses[sci].CourseCodes = append(slot.Courses[sci].CourseCodes, activeCodes...)
					slot.Courses[sci].Departments = append(slot.Courses[sci].Departments, activeDepartments...)
					slot.Courses[sci].Semesters = append(slot.Courses[sci].Semesters, activeSemesters...)
					slot.Courses[sci].StudentSets = append(slot.Courses[sci].StudentSets, course.StudentSet)
					slot.Courses[sci].Strength += course.Strength
					clubbed = true
					break
				}
			}
			if !clubbed {
				slot.Courses = append(slot.Courses, SlotCourse{
					CourseCodes: activeCodes,
					CourseName:  course.CourseName,
					Strength:    course.Strength,
					Departments: activeDepartments,
					Semesters:   activeSemesters,
					StudentSets: []map[string]bool{course.StudentSet},
					ClashSets:   []map[string]bool{course.ClashSet},
					Type:        targetType,
				})
			}

			totalScheduled++
			
			// Commit course-code scheduling state
			for code, baseSem := range course.CodeToSem {
				if targetType == "Regular" {
					if validSemSet[baseSem] {
						regScheduled[code] = true
						arrScheduled[code] = true
					} else {
						arrScheduled[code] = true
					}
				} else {
					arrScheduled[code] = true
				}
			}
			// Commit students placed in THIS course to perCourseScheduled
			if perCourseScheduled[courseKey] == nil {
				perCourseScheduled[courseKey] = make(map[string]bool)
			}
			for s := range dynRoster {
				perCourseScheduled[courseKey][s] = true
			}

			// --- STANDALONE LOGIC ---
			// If this course is an "Extra Elective" from a leader department, 
			// the user wants it to be scheduled ALONE.
			// Currently, we don't have the 'isExtra' flag here (it's internal to sort).
			// However, we can detect it: if the result of sort put it at the top 
			// AND it is an Elective AND it is the FIRST course in the slot, 
			// and we know the mode is 4/5 (regular heavy), we stop here.
			if targetType == "Regular" && course.CourseType == "Elective" && len(slot.Courses) == 1 {
				// Re-verify if this was an 'Extra' course?
				// To keep it simple and robust, we check if this department is still a leader.
				curDeptCount := 0
				for _, c := range pool {
					if !isCourseScheduled(c, targetType, regScheduled, arrScheduled) && c.Department == course.Department {
						curDeptCount++
					}
				}
				// If this dept has more courses left than others, it's a leader.
				isLeader := false
				for _, c := range pool {
					if !isCourseScheduled(c, targetType, regScheduled, arrScheduled) && c.Department != course.Department {
						otherCount := 0
						for _, c2 := range pool {
							if !isCourseScheduled(c2, targetType, regScheduled, arrScheduled) && c2.Department == c.Department {
								otherCount++
							}
						}
						if curDeptCount > otherCount {
							isLeader = true
							break
						}
					}
				}
				if isLeader {
					break // Schedule this Elective ALONE in this session
				}
			}
		}

		// --- Pass 2: Exclusive 5th Semester Gap Filling (Mode 5) ---
		// If Mode 5 AND NO COURSES were scheduled in Pass 1,
		// then S5 can take this session exclusively (Any morning or afternoon).
		if config.Mode == 5 && config.ExtraSem > 0 && len(slot.Courses) == 0 {
			// Try to fill from the correct pool (ExtraType)
			var extraPool []CourseInfo
			if config.ExtraType == "Regular" {
				extraPool = regCourses
			} else {
				extraPool = arrCourses
			}

			for i := 0; i < len(extraPool); i++ {
				course := extraPool[i]
				if isCourseScheduled(course, config.ExtraType, regScheduled, arrScheduled) {
					continue
				}
				isExtra := false
				for _, sem := range course.Semesters {
					if sem == config.ExtraSem {
						isExtra = true
						break
					}
				}
				if !isExtra {
					continue
				}

				// -- DYNAMIC ROSTER RECONSTRUCTION --
				dynRoster := make(map[string]bool)
				extraCourseKey := strings.ToUpper(strings.TrimSpace(course.CourseName))
				extraAlreadyIn := perCourseScheduled[extraCourseKey]
				extraCodesMap := make(map[string]bool)

				for code, baseSem := range course.CodeToSem {
					codeContributed := false
					if config.ExtraType == "Regular" {
						if baseSem == config.ExtraSem {
							if !regScheduled[code] {
								for s := range course.CodeToGlobal[code] {
									if extraAlreadyIn[s] { continue }
									if nativeSem, isRegular := studentNativeSem[s]; isRegular {
										if nativeSem != config.ExtraSem { continue }
									}
									dynRoster[s] = true
									codeContributed = true
								}
							}
						} else {
							if !arrScheduled[code] {
								for s := range course.CodeToArrear[code] {
									if !extraAlreadyIn[s] { 
										dynRoster[s] = true 
										codeContributed = true
									}
								}
							}
						}
					} else {
						if !arrScheduled[code] {
							for s := range course.CodeToArrear[code] {
								if !extraAlreadyIn[s] { 
									dynRoster[s] = true 
									codeContributed = true
								}
							}
						}
					}
					if codeContributed {
						extraCodesMap[code] = true
					}
				}

				if len(dynRoster) == 0 {
					continue
				}

				var activeExtraCodes []string
				extraSemMap := make(map[int]bool)
				extraDeptMap := make(map[string]bool)
				for code := range extraCodesMap {
					activeExtraCodes = append(activeExtraCodes, code)
					if sem, ok := course.CodeToSem[code]; ok {
						extraSemMap[sem] = true
					}
					if dept, ok := course.CodeToDept[code]; ok {
						for _, d := range strings.Split(dept, ",") {
							extraDeptMap[strings.TrimSpace(d)] = true
						}
					}
				}
				
				var activeExtraSemesters []int
				for sem := range extraSemMap {
					activeExtraSemesters = append(activeExtraSemesters, sem)
				}
				var activeExtraDepartments []string
				for dept := range extraDeptMap {
					activeExtraDepartments = append(activeExtraDepartments, dept)
				}

				if len(activeExtraCodes) == 0 {
					activeExtraCodes = strings.Split(course.CourseCode, "+")
					activeExtraSemesters = course.Semesters
					activeExtraDepartments = strings.Split(course.Department, ", ")
				}

				course.StudentSet = dynRoster
				course.Strength = len(dynRoster)

				// Check for student clashes with already scheduled courses in this slot
				if !canClubWith(course, slot.Courses) {
					continue
				}

				// Place or Club
				clubbed := false
				for sci := range slot.Courses {
					if strings.EqualFold(slot.Courses[sci].CourseName, course.CourseName) {
						slot.Courses[sci].CourseCodes = append(slot.Courses[sci].CourseCodes, activeExtraCodes...)
						slot.Courses[sci].Departments = append(slot.Courses[sci].Departments, activeExtraDepartments...)
						slot.Courses[sci].Semesters = append(slot.Courses[sci].Semesters, activeExtraSemesters...)
						slot.Courses[sci].StudentSets = append(slot.Courses[sci].StudentSets, course.StudentSet)
						slot.Courses[sci].Strength += course.Strength
						clubbed = true
						break
					}
				}

				if !clubbed {
					slot.Courses = append(slot.Courses, SlotCourse{
						CourseCodes: activeExtraCodes,
						CourseName:  course.CourseName,
						Strength:    course.Strength,
						Departments: activeExtraDepartments,
						Semesters:   activeExtraSemesters,
						StudentSets: []map[string]bool{course.StudentSet},
						ClashSets:   []map[string]bool{course.ClashSet},
						Type:        config.ExtraType,
					})
				}
				
				totalScheduled++

				// Commit course-code state
				for code, baseSem := range course.CodeToSem {
					if config.ExtraType == "Regular" {
						if baseSem == config.ExtraSem {
							regScheduled[code] = true
							arrScheduled[code] = true
						} else {
							arrScheduled[code] = true
						}
					} else {
						arrScheduled[code] = true
					}
				}
				// Commit students placed in THIS course
				if perCourseScheduled[extraCourseKey] == nil {
					perCourseScheduled[extraCourseKey] = make(map[string]bool)
				}
				for s := range dynRoster {
					perCourseScheduled[extraCourseKey][s] = true
				}
			}
		}

		result = append(result, slot)
	}

	return result, nil
}

// isCourseScheduled evaluates if a merged course has been fully satisfied
// based on each of its component codes individually.
func isCourseScheduled(c CourseInfo, targetType string, regScheduled map[string]bool, arrScheduled map[string]bool) bool {
	for code := range c.CodeToSem {
		if targetType == "Regular" {
			if !regScheduled[code] { return false }
		} else {
			if !arrScheduled[code] { return false }
		}
	}
	return true
}
