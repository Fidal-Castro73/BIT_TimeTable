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
	StudentSet   map[string]bool
}

// SlotCourse represents a single (possibly clubbed) course entry in a slot
type SlotCourse struct {
	CourseCodes []string // multiple if clubbed
	CourseName  string
	Strength    int
	Departments []string
	Semesters   []int
	StudentSets []map[string]bool // one per code
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

		// Load student set for clash detection
		studentSet, err := clash.GetStudentSet(db, c.CourseCode, regulationID, uploadType)
		if err == nil {
			c.StudentSet = studentSet
		}
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
		for _, existingSet := range sc.StudentSets {
			if clash.HasClash(candidate.StudentSet, existingSet) {
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

	// Track which courses are scheduled
	regScheduled := make([]bool, len(regCourses))
	arrScheduled := make([]bool, len(arrCourses))
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
		var scheduled []bool
		if targetType == "Regular" {
			pool = regCourses
			scheduled = regScheduled
		} else {
			pool = arrCourses
			scheduled = arrScheduled
		}

		// --- Pass 1: Primary Semesters (S1-S4) ---
		for i, course := range pool {
			if scheduled[i] {
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

			if !canClubWith(course, slot.Courses) {
				continue
			}

			// Clubbing logic...
			clubbed := false
			for sci := range slot.Courses {
				if strings.EqualFold(slot.Courses[sci].CourseName, course.CourseName) {
					slot.Courses[sci].CourseCodes = append(slot.Courses[sci].CourseCodes, course.CourseCode)
					slot.Courses[sci].Departments = append(slot.Courses[sci].Departments, course.Department)
					slot.Courses[sci].Semesters = append(slot.Courses[sci].Semesters, course.Semesters...)
					slot.Courses[sci].StudentSets = append(slot.Courses[sci].StudentSets, course.StudentSet)
					slot.Courses[sci].Strength += course.Strength
					clubbed = true
					break
				}
			}
			if !clubbed {
				slot.Courses = append(slot.Courses, SlotCourse{
					CourseCodes: []string{course.CourseCode},
					CourseName:  course.CourseName,
					Strength:    course.Strength,
					Departments: []string{course.Department},
					Semesters:   course.Semesters,
					StudentSets: []map[string]bool{course.StudentSet},
				})
			}

			scheduled[i] = true
			totalScheduled++
			
			// Commit state back (Go slices refs...)
			if targetType == "Regular" {
				regScheduled[i] = true
			} else {
				arrScheduled[i] = true
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
				for idx, c := range pool {
					if !scheduled[idx] && c.Department == course.Department {
						curDeptCount++
					}
				}
				// If this dept has more courses left than others, it's a leader.
				isLeader := false
				for idx, c := range pool {
					if !scheduled[idx] && c.Department != course.Department {
						otherCount := 0
						for k, c2 := range pool { if !scheduled[k] && c2.Department == c.Department { otherCount++ } }
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
			var extraScheduled []bool
			if config.ExtraType == "Regular" {
				extraPool = regCourses
				extraScheduled = regScheduled
			} else {
				extraPool = arrCourses
				extraScheduled = arrScheduled
			}

			for i, course := range extraPool {
				if extraScheduled[i] {
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

				// Check for student clashes with already scheduled courses in this slot
				if !canClubWith(course, slot.Courses) {
					continue
				}

				// Place or Club
				clubbed := false
				for sci := range slot.Courses {
					if strings.EqualFold(slot.Courses[sci].CourseName, course.CourseName) {
						slot.Courses[sci].CourseCodes = append(slot.Courses[sci].CourseCodes, course.CourseCode)
						slot.Courses[sci].Departments = append(slot.Courses[sci].Departments, course.Department)
						slot.Courses[sci].Semesters = append(slot.Courses[sci].Semesters, course.Semesters...)
						slot.Courses[sci].StudentSets = append(slot.Courses[sci].StudentSets, course.StudentSet)
						slot.Courses[sci].Strength += course.Strength
						clubbed = true
						break
					}
				}

				if !clubbed {
					slot.Courses = append(slot.Courses, SlotCourse{
						CourseCodes: []string{course.CourseCode},
						CourseName:  course.CourseName,
						Strength:    course.Strength,
						Departments: []string{course.Department},
						Semesters:   course.Semesters,
						StudentSets: []map[string]bool{course.StudentSet},
					})
				}
				
				extraScheduled[i] = true
				totalScheduled++
				if config.ExtraType == "Regular" {
					regScheduled[i] = true
				} else {
					arrScheduled[i] = true
				}
			}
		}

		result = append(result, slot)
	}

	return result, nil
}
