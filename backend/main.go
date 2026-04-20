package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"bittimetable/internal/engine"
	"bittimetable/internal/pdf"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
)

var db *sql.DB

func initDB() {
	var err error

	// Open connection to standard XAMPP MySQL root (no password by default)
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/")
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}

	// Create database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS bit_timetable")
	if err != nil {
		log.Fatal("Failed to create database:", err)
	}
	db.Close()

	// Connect to the specific database
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/bit_timetable")
	if err != nil {
		log.Fatal("Failed to connect to bit_timetable:", err)
	}

	// Create table securely with the requested columns
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS student_data (
		id INT AUTO_INCREMENT PRIMARY KEY,
		register_no VARCHAR(50),
		student_name VARCHAR(255),
		course_name VARCHAR(255),
		course_code VARCHAR(50),
		semester INT,
		course_type ENUM('Core', 'Elective', 'Honors', 'Minors', 'Open Elective'),
		department VARCHAR(50),
		batch INT,
		regulation_id INT,
		upload_type VARCHAR(50) -- Supports 'Regular', 'Arrear', 'Mixed'
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Create timetables table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS timetables (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255),
		regulation_id INT,
		upload_type VARCHAR(50), 
		day_config JSON,
		status ENUM('pending', 'approved') DEFAULT 'pending',
		generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Fatal("Error creating timetables table:", err)
	}

	// Create timetable_slots table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS timetable_slots (
		id INT AUTO_INCREMENT PRIMARY KEY,
		timetable_id INT,
		day_number INT,
		session ENUM('FN','AN'),
		course_codes TEXT,
		course_name VARCHAR(255),
		strength INT,
		departments TEXT,
		semesters TEXT,
		FOREIGN KEY (timetable_id) REFERENCES timetables(id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal("Error creating timetable_slots table:", err)
	}

	// Dynamic Schema Update: Ensure columns exist if the table was created previously
	db.Exec("ALTER TABLE timetable_slots ADD COLUMN semesters TEXT AFTER departments")
	db.Exec("ALTER TABLE timetable_slots ADD COLUMN session_type VARCHAR(50) AFTER semesters")
	db.Exec("ALTER TABLE timetable_slots ADD COLUMN student_rolls JSON AFTER session_type")

	// 11. Database Deep Scrub (One-time Migration)
	// Optimized: Only run on rows that actually need cleaning to save time on startup
	_, _ = db.Exec("UPDATE student_data SET course_code = REPLACE(REPLACE(REPLACE(course_code, '-', ''), '*', ''), ' ', '') WHERE course_code LIKE '%-%' OR course_code LIKE '%*%' OR course_code LIKE '% %'")

	fmt.Println("MySQL Database & Schema Initialized Successfully!")
}


// Extract Core, Elective, Honors, Minors
func parseCourseType(courseCode string) string {
	courseCode = strings.TrimSpace(courseCode)

	// Rule 1: Priority Check (3rd and 5th letters)
	// Example: 21OBT01 -> 3rd (index 2) is 'O'
	if len(courseCode) >= 3 && (string(courseCode[2]) == "O" || string(courseCode[2]) == "o") {
		// Minimum length check to safely access 5th letter (index 4) if needed
		if len(courseCode) >= 5 {
			fifthChar := strings.ToUpper(string(courseCode[4]))
			if fifthChar == "H" {
				return "Honors"
			}
			if fifthChar == "M" {
				return "Minors"
			}
		}
		// If 3rd is 'O' but 5th isn't H/M, it's an Open Elective
		return "Open Elective"
	}

	// Rule 2: Standard Fallback (Last 3 digits check)
	// Example: 18BT101 -> last 3 is '101'
	if len(courseCode) >= 3 {
		last3 := courseCode[len(courseCode)-3:]
		firstDigitOfLast3 := string(last3[0])

		if firstDigitOfLast3 >= "1" && firstDigitOfLast3 <= "8" {
			return "Core"
		} else if firstDigitOfLast3 == "0" || strings.ToUpper(firstDigitOfLast3) == "O" {
			return "Elective"
		}
	}

	return "Core" // Final default
}

// Convert Semester Roman Strings to Integers
func parseSemester(sem string) int {
	switch strings.ToUpper(strings.TrimSpace(sem)) {
	case "I":
		return 1
	case "II":
		return 2
	case "III":
		return 3
	case "IV":
		return 4
	case "V":
		return 5
	case "VI":
		return 6
	case "VII":
		return 7
	case "VIII":
		return 8
	}
	if val, err := strconv.Atoi(sem); err == nil {
		return val
	}
	return 0
}

// Smart Parser for Batch and Department
func parseBatchAndDept(regNo string) (int, string) {
	regNo = strings.TrimSpace(regNo)

	// Reg No starts with 7376
	// Format: 7376 (College) + 21 (Batch) + 1 (Level?) + CS (Dept) + ...
	regex7376 := regexp.MustCompile(`^7376(\d{2})\d?([A-Za-z]{2,3})`)
	matches := regex7376.FindStringSubmatch(regNo)

	if len(matches) >= 3 {
		batchSuffix, _ := strconv.Atoi(matches[1])
		batch := 2000 + batchSuffix
		dept := strings.ToUpper(matches[2])
		return batch, dept
	}

	// Old Format: 18 (Batch) + 2 (Level?) + BT (Dept) + ...
	regexOld := regexp.MustCompile(`^(\d{2})\d?([A-Za-z]{2,3})`)
	matchesOld := regexOld.FindStringSubmatch(regNo)

	if len(matchesOld) >= 3 {
		batchSuffix, _ := strconv.Atoi(matchesOld[1])
		batch := 2000 + batchSuffix
		dept := strings.ToUpper(matchesOld[2])
		return batch, dept
	}

	return 2000, "UNKNOWN"
}

func uploadHandler(c *gin.Context) {
	regulationStr := c.PostForm("regulation_id")
	uploadType := c.PostForm("upload_type") // 'Regular' or 'Arrear'

	regulationID, err := strconv.Atoi(regulationStr)
	if err != nil || (uploadType != "Regular" && uploadType != "Arrear") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid regulation_id or upload_type"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file received"})
		return
	}

	// Save temporary file in system temp
	tempFilePath := "temp_" + file.Filename
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Unable to save file to server memory"})
		return
	}

	// Open Excel file
	f, err := excelize.OpenFile(tempFilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Unable to open excel file. Ensure it's a valid .xlsx file."})
		return
	}
	defer f.Close()

	sheetName := f.GetSheetList()[0] // Get first sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Unable to read rows from the Excel sheet"})
		return
	}

	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "The Excel file is empty or missing data rows"})
		return
	}

	// ----------------------------------------------------
	// RULE 2: Delete-Before-Insert Strategy
	// ----------------------------------------------------
	_, err = db.Exec("DELETE FROM student_data WHERE regulation_id = ? AND upload_type = ?", regulationID, uploadType)
	if err != nil {
		log.Println("Delete Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to verify database safety rules"})
		return
	}

	// Prepare dynamic SQL insert
	stmt, err := db.Prepare(`
		INSERT INTO student_data 
		(register_no, student_name, course_name, course_code, semester, course_type, department, batch, regulation_id, upload_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to prepare SQL insertion"})
		return
	}
	defer stmt.Close()

	insertedCount := 0

	// Loop through rows
	for i, row := range rows {
		if i == 0 { // Skip header
			continue
		}

		// Fill missing columns in row array if empty at the end
		for len(row) < 7 {
			row = append(row, "")
		}

		// Columns mapping based on requested Excel sample format
		// 0: S.no, 1: Course Code, 2: Course Name, 3: Semester, 4: Reg No, 5: Student Name
		courseCode := strings.TrimSpace(row[1])
		courseName := strings.TrimSpace(row[2])
		semStr := strings.TrimSpace(row[3])
		regNo := strings.TrimSpace(row[4])
		studentName := strings.TrimSpace(row[5])

		// Basic validation skip
		if regNo == "" || courseCode == "" {
			continue
		}

		// PARSERS
		semester := parseSemester(semStr)
		batch, dept := parseBatchAndDept(regNo)
		// Alphanumeric Clean for Storage
		regAlphanum := regexp.MustCompile(`[^a-zA-Z0-9]`)
		cleanCourseCode := regAlphanum.ReplaceAllString(courseCode, "")
		courseType := parseCourseType(cleanCourseCode)

		_, err = stmt.Exec(
			regNo, studentName, courseName, cleanCourseCode, semester,
			courseType, dept, batch, regulationID, uploadType,
		)

		if err != nil {
			log.Printf("DB Error on row %d (RegNo: %s): %v", i+1, regNo, err)
			continue
		}

		insertedCount++
	}

	// Success response back to frontend
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("%d %s records processed & stored successfully for Regulation %d!", insertedCount, uploadType, regulationID),
		"inserted": insertedCount,
	})
}

func getImportsHandler(c *gin.Context) {
	rows, err := db.Query(`
		SELECT regulation_id, upload_type, COUNT(*) as records_count 
		FROM student_data 
		GROUP BY regulation_id, upload_type
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch import history"})
		return
	}
	defer rows.Close()

	type ImportHistory struct {
		Regulation   int    `json:"regulation"`
		Category     string `json:"category"`
		RecordsCount int    `json:"recordsCount"`
		Status       string `json:"status"`
	}

	imports := []ImportHistory{}
	for rows.Next() {
		var item ImportHistory
		if err := rows.Scan(&item.Regulation, &item.Category, &item.RecordsCount); err != nil {
			continue
		}
		item.Status = "imported"
		imports = append(imports, item)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "imports": imports})
}

func deleteImportHandler(c *gin.Context) {
	regulationID, _ := strconv.Atoi(c.Query("regulation_id"))
	uploadType := c.Query("upload_type")

	if regulationID == 0 || uploadType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing params"})
		return
	}

	_, err := db.Exec("DELETE FROM student_data WHERE regulation_id = ? AND upload_type = ?", regulationID, uploadType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Import removed from database successfully"})
}

func updateImportHandler(c *gin.Context) {
	var body struct {
		OldRegulation int    `json:"old_regulation"`
		OldCategory   string `json:"old_category"`
		NewRegulation int    `json:"new_regulation"`
		NewCategory   string `json:"new_category"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid JSON"})
		return
	}

	_, err := db.Exec(`
		UPDATE student_data 
		SET regulation_id = ?, upload_type = ? 
		WHERE regulation_id = ? AND upload_type = ?`,
		body.NewRegulation, body.NewCategory, body.OldRegulation, body.OldCategory)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Records updated successfully"})
}

// --- Timetable Generation Handler ---

func generateHandler(c *gin.Context) {
	var body struct {
		Name         string           `json:"name"`
		RegulationID int              `json:"regulation_id"`
		DayConfig    engine.DayConfig `json:"day_config"`
		SortOptions  engine.SortOptions `json:"sort_options"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	// Guard: Check if student data exists for this regulation
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM student_data WHERE regulation_id = ?", body.RegulationID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false, 
			"error": fmt.Sprintf("No student data found for Regulation Year %d. Please upload data first.", body.RegulationID),
		})
		return
	}

	slots, err := engine.Generate(db, body.RegulationID, body.DayConfig, body.SortOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Save timetable to DB
	configJSON, _ := json.Marshal(body.DayConfig)
	// Use the first session's type or Mixed
	labelType := body.DayConfig.OddFNType
	if labelType == "" { labelType = "Mixed" }

	res, err := db.Exec(`INSERT INTO timetables (name, regulation_id, upload_type, day_config, status) VALUES (?, ?, ?, ?, 'pending')`,
		body.Name, body.RegulationID, labelType, string(configJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save timetable"})
		return
	}
	timetableID, _ := res.LastInsertId()

	// Save slots
	for _, slot := range slots {
		for _, sc := range slot.Courses {
			codes := strings.Join(sc.CourseCodes, ",")
			depts := strings.Join(sc.Departments, ",")
			var semBatch []string
			for _, s := range sc.Semesters {
				semBatch = append(semBatch, strconv.Itoa(s))
			}
			sems := strings.Join(semBatch, ",")

			// Extract flat list of unique student rolls assigned by the dynamic roster reconstructor
			var rollList []string
			rollMap := make(map[string]bool)
			for _, m := range sc.StudentSets {
				for roll := range m {
					if !rollMap[roll] {
						rollMap[roll] = true
						rollList = append(rollList, roll)
					}
				}
			}
			rollsJSON, _ := json.Marshal(rollList)

			_, err = db.Exec(`INSERT INTO timetable_slots (timetable_id, day_number, session, course_codes, course_name, strength, departments, semesters, session_type, student_rolls) VALUES (?,?,?,?,?,?,?,?,?,?)`,
				timetableID, slot.DayNumber, slot.Session, codes, sc.CourseName, sc.Strength, depts, sems, sc.Type, string(rollsJSON))
			if err != nil {
				log.Println("Slot insert error:", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"timetable_id": timetableID,
		"slots":        slots,
		"message":      fmt.Sprintf("Generated %d sessions successfully", len(slots)),
	})
}

// --- Dashboard Stats Handler ---

func getDashboardStats(c *gin.Context) {
	timetableID, _ := strconv.Atoi(c.Query("timetable_id"))
	regID, _ := strconv.Atoi(c.Query("regulation_id"))
	uploadType := c.Query("upload_type")
	semestersStr := c.Query("semesters")

	// Prepare semester filter safely
	semFilter := ""
	if semestersStr != "" {
		parts := strings.Split(semestersStr, ",")
		var validSems []string
		for _, p := range parts {
			if _, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				validSems = append(validSems, strings.TrimSpace(p))
			}
		}
		if len(validSems) > 0 {
			semFilter = " AND semester IN (" + strings.Join(validSems, ",") + ")"
		}
	}

	excludeFilter := ` AND LOWER(course_name) NOT LIKE '%laboratory' AND LOWER(course_name) NOT LIKE '%project%'`

	// Stats logic
	var totalStrength int
	var uniqueCourses int
	typeDistrib := map[string]int{}

	if timetableID > 0 {
		// --- REFLECT OUTPUT (Based on timetable_slots) ---
		
		// Total Strength (Sum of strengths of all generated slots)
		db.QueryRow(`SELECT COALESCE(SUM(strength), 0) FROM timetable_slots WHERE timetable_id=?`, timetableID).Scan(&totalStrength)
		
		// Unique Courses (Count slots)
		db.QueryRow(`SELECT COUNT(*) FROM timetable_slots WHERE timetable_id=?`, timetableID).Scan(&uniqueCourses)
		
		// Type Distribution (Join with student_data)
		typeRows, _ := db.Query(`
			SELECT sd.course_type, COUNT(DISTINCT ts.course_name) 
			FROM timetable_slots ts
			JOIN student_data sd ON ts.course_name = sd.course_name
			WHERE ts.timetable_id=?
			GROUP BY sd.course_type`, timetableID)
		if typeRows != nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var ct string
				var cnt int
				typeRows.Scan(&ct, &cnt)
				typeDistrib[ct] = cnt
			}
		}
	} else {
		// --- REFLECT INPUT (Based on database) ---
		if regID == 0 || uploadType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing params"})
			return
		}

		db.QueryRow(`SELECT COUNT(DISTINCT register_no) FROM student_data WHERE regulation_id=? AND upload_type=?`+semFilter+excludeFilter, regID, uploadType).Scan(&totalStrength)
		db.QueryRow(`SELECT COUNT(DISTINCT course_name) FROM student_data WHERE regulation_id=? AND upload_type=?`+semFilter+excludeFilter, regID, uploadType).Scan(&uniqueCourses)

		typeRows, _ := db.Query(`SELECT course_type, COUNT(DISTINCT course_name) as cnt FROM student_data WHERE regulation_id=? AND upload_type=?`+semFilter+excludeFilter+` GROUP BY course_type`, regID, uploadType)
		if typeRows != nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var ct string
				var cnt int
				typeRows.Scan(&ct, &cnt)
				typeDistrib[ct] = cnt
			}
		}
	}

	// Dept breakdown
	type DeptStat struct {
		Department string `json:"department"`
		Count      int    `json:"count"`
	}
	var deptStats []DeptStat
	
	if timetableID > 0 {
		// Aggregate departments manually because they are saved as comma-separated strings
		rows, _ := db.Query(`SELECT departments FROM timetable_slots WHERE timetable_id=?`, timetableID)
		if rows != nil {
			defer rows.Close()
			deptMap := make(map[string]int)
			for rows.Next() {
				var depts string
				rows.Scan(&depts)
				parts := strings.Split(depts, ",")
				seenInThisSlot := make(map[string]bool)
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" || seenInThisSlot[p] {
						continue
					}
					deptMap[p]++
					seenInThisSlot[p] = true
				}
			}
			for d, count := range deptMap {
				deptStats = append(deptStats, DeptStat{Department: d, Count: count})
			}
		}
	} else {
		rows, _ := db.Query(`SELECT department, COUNT(DISTINCT course_name) as cnt FROM student_data WHERE regulation_id=? AND upload_type=?`+semFilter+excludeFilter+` GROUP BY department ORDER BY cnt DESC`, regID, uploadType)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var ds DeptStat
				rows.Scan(&ds.Department, &ds.Count)
				deptStats = append(deptStats, ds)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"total_strength":  totalStrength,
		"unique_courses": uniqueCourses,
		"type_distrib":   typeDistrib,
		"dept_stats":     deptStats,
	})
}

// --- Get All Timetables ---

func getTimetablesHandler(c *gin.Context) {
	status := c.DefaultQuery("status", "approved")
	query := `SELECT id, name, regulation_id, upload_type, status, day_config, generated_at FROM timetables WHERE status=? ORDER BY generated_at DESC`
	args := []interface{}{status}

	if status == "all" {
		query = `SELECT id, name, regulation_id, upload_type, status, day_config, generated_at FROM timetables ORDER BY generated_at DESC`
		args = []interface{}{}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "DB error: " + err.Error()})
		return
	}
	defer rows.Close()

	type TT struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		RegulationID int    `json:"regulation_id"`
		UploadType   string `json:"upload_type"`
		Status       string `json:"status"`
		DayConfig    string `json:"day_config"`
		GeneratedAt  string `json:"generated_at"`
	}
	var tts []TT
	for rows.Next() {
		var tt TT
		rows.Scan(&tt.ID, &tt.Name, &tt.RegulationID, &tt.UploadType, &tt.Status, &tt.DayConfig, &tt.GeneratedAt)
		tts = append(tts, tt)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timetables": tts})
}

// --- Get Single Timetable with Slots ---

func getTimetableHandler(c *gin.Context) {
	id := c.Param("id")

	// 1. Get Timetable Metadata
	var tt struct {
		ID           int             `json:"id"`
		Name         string          `json:"name"`
		RegulationID int             `json:"regulation_id"`
		UploadType   string          `json:"upload_type"`
		DayConfig    json.RawMessage `json:"day_config"`
		Status       string          `json:"status"`
	}
	err := db.QueryRow(`SELECT id, name, regulation_id, upload_type, day_config, status FROM timetables WHERE id=?`, id).
		Scan(&tt.ID, &tt.Name, &tt.RegulationID, &tt.UploadType, &tt.DayConfig, &tt.Status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Timetable metadata not found"})
		return
	}

	// 2. Get Slots
	rows, err := db.Query(`SELECT id, day_number, session, course_codes, course_name, strength, departments, semesters FROM timetable_slots WHERE timetable_id=? ORDER BY day_number, session`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load slots"})
		return
	}
	defer rows.Close()

	type Slot struct {
		ID          int    `json:"id"`
		DayNumber   int    `json:"day_number"`
		Session     string `json:"session"`
		CourseCodes string `json:"course_codes"`
		CourseName  string `json:"course_name"`
		Strength    int    `json:"strength"`
		Departments string `json:"departments"`
		Semesters   string `json:"semesters"`
	}
	var slots []Slot
	for rows.Next() {
		var s Slot
		rows.Scan(&s.ID, &s.DayNumber, &s.Session, &s.CourseCodes, &s.CourseName, &s.Strength, &s.Departments, &s.Semesters)
		slots = append(slots, s)
	}

	// 3. Optimized Global Clash Detection for UI Hints
	clashingIDs := make(map[int]bool)
	
	// Group slots by Day + Session
	sessionMap := make(map[string][]Slot)
	for _, s := range slots {
		key := fmt.Sprintf("%d-%s", s.DayNumber, s.Session)
		sessionMap[key] = append(sessionMap[key], s)
	}

	for _, sSlots := range sessionMap {
		// Even if there's only 1 card in a session, it could have internal clashes (clubbed courses)
		// so we must check all sessions.

		// 1. Gather all course codes in this session
		var sessionCodes []string
		for _, s := range sSlots {
			sessionCodes = append(sessionCodes, strings.Split(s.CourseCodes, ",")...)
		}

		// 2. Fetch all student records for these codes in one go
		placeholders := make([]string, len(sessionCodes))
		args := []interface{}{tt.RegulationID}
		for i, c := range sessionCodes { placeholders[i] = "?"; args = append(args, strings.TrimSpace(c)) }

		q := fmt.Sprintf(`SELECT register_no, course_code FROM student_data WHERE regulation_id=? AND course_code IN (%s)`, strings.Join(placeholders, ","))
		rows, err := db.Query(q, args...)
		if err != nil { continue }

		// 3. Map students to slot IDs
		// studentToSlots[regNo] = [slotID1, slotID2...]
		studentCountsInSession := make(map[string]int)
		studentToSlots := make(map[string][]int)
		
		for rows.Next() {
			var reg, code string
			rows.Scan(&reg, &code)
			cleanCode := strings.TrimSpace(code)
			studentCountsInSession[reg]++
			
			for _, slot := range sSlots {
				// Use exact match check for codes within comma-sep string
				found := false
				for _, c := range strings.Split(slot.CourseCodes, ",") {
					if strings.TrimSpace(c) == cleanCode {
						found = true
						break
					}
				}
				if found {
					studentToSlots[reg] = append(studentToSlots[reg], slot.ID)
				}
			}
		}
		rows.Close()

		// 4. Identify clashing slot IDs
		for reg, count := range studentCountsInSession {
			if count > 1 {
				fmt.Printf("   [DASHBOARD-CLASH] Student %s found %d times in session. Flagging slots: %v\n", reg, count, studentToSlots[reg])
				for _, sid := range studentToSlots[reg] {
					clashingIDs[sid] = true
				}
			}
		}
	}

	if len(clashingIDs) > 0 {
		fmt.Printf("🛡️ UI FEEDBACK: Flagging %d clashing slot(s) for visual glow\n", len(clashingIDs))
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timetable": tt,
		"slots":     slots,
		"clashing_ids": clashingIDs,
	})
}

// --- Move Course Handler (Drag-and-Drop) ---

func checkMoveConflictsHandler(c *gin.Context) {
	id := c.Param("id")
	slotID := c.Param("slotId")
	var body struct {
		NewDay     int    `json:"new_day"`
		NewSession string `json:"new_session"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid body"})
		return
	}

	// 1. Get the regulation ID and type of this timetable
	var regID int
	var ut string
	err := db.QueryRow("SELECT regulation_id, upload_type FROM timetables WHERE id=?", id).Scan(&regID, &ut)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Timetable not found"})
		return
	}

	// 2. Get the course code(s) of the slot being moved
	var moveCodes string
	err = db.QueryRow("SELECT course_codes FROM timetable_slots WHERE id=?", slotID).Scan(&moveCodes)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Slot not found"})
		return
	}
	moveList := strings.Split(moveCodes, ",")

	// 3. Get all course codes already in the target session (excluding itself if it's already there)
	rows, err := db.Query("SELECT course_codes FROM timetable_slots WHERE timetable_id=? AND day_number=? AND session=? AND id <> ?", id, body.NewDay, body.NewSession, slotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "DB error query"})
		return
	}
	defer rows.Close()

	var targetCodes []string
	for rows.Next() {
		var codes string
		rows.Scan(&codes)
		targetCodes = append(targetCodes, strings.Split(codes, ",")...)
	}

	if len(targetCodes) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "conflicts": []string{}})
		return
	}

	// 4. Find students who exist in both sets (Clash detection)
	// We check which students from the MOVE set are also in the TARGET set
	placeholdersMove := make([]string, len(moveList))
	argsMove := []interface{}{regID, ut}
	for i, c := range moveList { placeholdersMove[i] = "?"; argsMove = append(argsMove, strings.TrimSpace(c)) }

	placeholdersTarget := make([]string, len(targetCodes))
	argsTarget := []interface{}{regID, ut}
	for i, c := range targetCodes { placeholdersTarget[i] = "?"; argsTarget = append(argsTarget, strings.TrimSpace(c)) }

	// Find intersection of Register Numbers
	conflictQuery := fmt.Sprintf(`
		SELECT DISTINCT student_name 
		FROM student_data 
		WHERE regulation_id = ? AND upload_type = ? AND course_code IN (%s)
		AND register_no IN (
			SELECT register_no 
			FROM student_data 
			WHERE regulation_id = ? AND upload_type = ? AND course_code IN (%s)
		)`, strings.Join(placeholdersMove, ","), strings.Join(placeholdersTarget, ","))

	finalArgs := append(argsMove, argsTarget...)
	crows, err := db.Query(conflictQuery, finalArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Conflict check failed: " + err.Error()})
		return
	}
	defer crows.Close()

	var clashingStudents []string
	for crows.Next() {
		var name string
		crows.Scan(&name)
		clashingStudents = append(clashingStudents, name)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "conflicts": clashingStudents})
}

func moveCourseHandler(c *gin.Context) {
	slotID := c.Param("slotId")
	var body struct {
		NewDay     int    `json:"new_day"`
		NewSession string `json:"new_session"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid body"})
		return
	}

	_, err := db.Exec(`UPDATE timetable_slots SET day_number=?, session=? WHERE id=?`,
		body.NewDay, body.NewSession, slotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Move failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Course moved successfully"})
}

// --- Delete Timetable ---

func deleteTimetableHandler(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec(`DELETE FROM timetables WHERE id=?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Timetable deleted"})
}

// --- Approve Timetable ---

func approveTimetableHandler(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec(`UPDATE timetables SET status='approved' WHERE id=?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Approval failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Timetable approved"})
}

// --- Revert Timetable to Pending ---

func revertTimetableHandler(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec(`UPDATE timetables SET status='pending' WHERE id=?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Revert failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Timetable reverted to pending"})
}

// --- PDF Export Handler ---

func getPDFHandler(c *gin.Context) {
	id := c.Param("id")

	// Load timetable metadata
	var name string
	db.QueryRow(`SELECT name FROM timetables WHERE id=?`, id).Scan(&name)

	// Load slots
	rows, err := db.Query(`SELECT id, day_number, session, course_codes, course_name, strength, departments FROM timetable_slots WHERE timetable_id=? ORDER BY day_number, session`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load slots"})
		return
	}
	defer rows.Close()

	var slots []engine.ScheduleSlot
	slotMap := map[string]*engine.ScheduleSlot{}

	for rows.Next() {
		var slotID, dayNumber, strength int
		var session, codes, courseName, departments string
		rows.Scan(&slotID, &dayNumber, &session, &codes, &courseName, &strength, &departments)

		key := fmt.Sprintf("%d-%s", dayNumber, session)
		if _, exists := slotMap[key]; !exists {
			slotMap[key] = &engine.ScheduleSlot{DayNumber: dayNumber, Session: session}
			slots = append(slots, *slotMap[key])
		}

		sc := engine.SlotCourse{
			CourseCodes: strings.Split(codes, ","),
			CourseName:  courseName,
			Strength:    strength,
			Departments: strings.Split(departments, ","),
		}

		for i := range slots {
			if slots[i].DayNumber == dayNumber && slots[i].Session == session {
				slots[i].Courses = append(slots[i].Courses, sc)
				break
			}
		}
	}

	pdfBytes, err := pdf.Export(slots, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="timetable.pdf"`)
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func getSessionStudentsHandler(c *gin.Context) {
	id := c.Param("id")
	day := c.Param("day")
	period := c.Param("period")

	// 1. Get Timetable Metadata
	var regID int
	var ut string
	var dayConfigJSON string
	err := db.QueryRow(`SELECT regulation_id, upload_type, day_config FROM timetables WHERE id=?`, id).Scan(&regID, &ut, &dayConfigJSON)
	if err != nil {
		fmt.Printf("❌ Error: Timetable %s not found in DB\n", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Timetable not found"})
		return
	}

	var config engine.DayConfig
	json.Unmarshal([]byte(dayConfigJSON), &config)

	// Compute validSemesters for the current slot
	var validSemesters []int
	dayInt, _ := strconv.Atoi(day)
	isOdd := (dayInt % 2) != 0
	if isOdd {
		if period == "FN" { validSemesters = config.OddFN } else { validSemesters = config.OddAN }
	} else {
		if period == "FN" { validSemesters = config.EvenFN } else { validSemesters = config.EvenAN }
	}
	if len(validSemesters) == 0 && config.Mode == 5 {
		validSemesters = []int{config.ExtraSem}
	}

	// 2. Get Courses and EXACT Roster for this session
	rows, err := db.Query(`
		SELECT course_codes, session_type, student_rolls FROM timetable_slots 
		WHERE timetable_id = ? AND day_number = ? AND session = ?
	`, id, day, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load session courses"})
		return
	}
	defer rows.Close()

	var allCodes []string
	var exactRolls []string
	var sessionType string
	for rows.Next() {
		var codes, sType string
		var rollsJSON sql.NullString
		rows.Scan(&codes, &sType, &rollsJSON)
		sessionType = sType
		
		if rollsJSON.Valid && rollsJSON.String != "" {
			var rolls []string
			if err := json.Unmarshal([]byte(rollsJSON.String), &rolls); err == nil {
				exactRolls = append(exactRolls, rolls...)
			}
		}

		if codes != "" {
			parts := strings.Split(codes, ",")
			for _, p := range parts {
				clean := strings.TrimSpace(strings.ToUpper(p))
				if clean != "" {
					allCodes = append(allCodes, clean)
				}
			}
		}
	}

	fmt.Printf("🔍 Session Details: Timetable=%s, Day=%s, Sess=%s | Reg=%d, Type=%s | Codes=%v\n", id, day, period, regID, ut, allCodes)

	if len(allCodes) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "students": []interface{}{}})
		return
	}

	// 3. Fetch Students using EXACT roster mapping
	args := []interface{}{regID}
	queryStr := `
		SELECT student_name, register_no, course_code, course_name, department, upload_type
		FROM student_data 
		WHERE regulation_id = ? 
	`

	if len(exactRolls) > 0 {
		rollPlaceholders := make([]string, len(exactRolls))
		for i, roll := range exactRolls {
			rollPlaceholders[i] = "?"
			args = append(args, roll)
		}
		queryStr += ` AND register_no IN (` + strings.Join(rollPlaceholders, ",") + `)`
	} else {
		// Fallback for older databases without student_rolls
		if sessionType == "Arrear" {
			queryStr += " AND upload_type = 'Arrear' "
		} else if sessionType == "Regular" {
			semStrs := []string{}
			for _, s := range validSemesters {
				semStrs = append(semStrs, strconv.Itoa(s))
			}
			if len(semStrs) > 0 {
				queryStr += fmt.Sprintf(" AND (upload_type = 'Arrear' OR (upload_type = 'Regular' AND semester IN (%s))) ", strings.Join(semStrs, ","))
			} else {
				queryStr += " AND upload_type = 'Arrear' "
			}
		}
	}

	codePlaceholders := make([]string, len(allCodes))
	for i, code := range allCodes {
		codePlaceholders[i] = "?"
		args = append(args, code)
	}

	queryStr += ` AND (TRIM(UPPER(course_code)) IN (` + strings.Join(codePlaceholders, ",") + `)`
	queryStr += ` OR TRIM(REPLACE(UPPER(course_code), '-', '')) IN (` + strings.Join(codePlaceholders, ",") + `))`
	
	// Must append codes again for the OR clause
	for _, code := range allCodes {
		args = append(args, code)
	}

	queryStr += ` ORDER BY register_no`

	studentRows, err := db.Query(queryStr, args...)
	if err != nil {
		fmt.Printf("❌ SQL Query Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		return
	}
	defer studentRows.Close()

	type StudentDetail struct {
		Name       string `json:"name"`
		Roll       string `json:"roll"`
		CourseCode string `json:"course_code"`
		CourseName string `json:"course_name"`
		Dept       string `json:"dept"`
		Type       string `json:"type"`
		IsClashing bool   `json:"is_clashing"`
	}
	var results []StudentDetail
	regCounts := make(map[string]int)
	for studentRows.Next() {
		var name, roll, code, cname, dept, utype sql.NullString
		if err := studentRows.Scan(&name, &roll, &code, &cname, &dept, &utype); err != nil {
			fmt.Printf("⚠️ Scan Skip: %v\n", err)
			continue
		}
		
		s := StudentDetail{
			Name:       name.String,
			Roll:       roll.String,
			CourseCode: code.String,
			CourseName: cname.String,
			Dept:       dept.String,
			Type:       utype.String,
		}
		results = append(results, s)
		regCounts[s.Roll]++
	}

	// Dynamic Clash Calculation
	uniqueClashes := make(map[string]bool)
	for i := range results {
		if regCounts[results[i].Roll] > 1 {
			results[i].IsClashing = true
			uniqueClashes[results[i].Roll] = true
		}
	}
	clashCount := len(uniqueClashes)

	// Sort by IsClashing (Descending: true comes before false)
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].IsClashing && !results[j].IsClashing {
			return true
		}
		if !results[i].IsClashing && results[j].IsClashing {
			return false
		}
		return strings.Compare(results[i].Roll, results[j].Roll) < 0
	})

	// Logging unique clashing student names
	if clashCount > 0 {
		fmt.Printf("⚠️ WARNING: FOUND %d UNIQUE CLASHING STUDENTS in Day %s Session %s\n", clashCount, day, period)
		for roll := range uniqueClashes {
			fmt.Printf("   - [CLASH] Roll: %s\n", roll)
		}
	}

	fmt.Printf("✅ Success: Timetable %s Session Found %d students (%d clashing)\n", id, len(results), clashCount)
	c.JSON(http.StatusOK, gin.H{"success": true, "students": results})
}

func main() {
	initDB()

	r := gin.Default()

	// CORS Protection - allows React to talk to Go
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// Import routes
	r.POST("/api/upload", uploadHandler)
	r.GET("/api/imports", getImportsHandler)
	r.DELETE("/api/delete", deleteImportHandler)
	r.POST("/api/update", updateImportHandler)

	// Timetable generation routes
	r.POST("/api/generate", generateHandler)
	r.GET("/api/timetables", getTimetablesHandler)
	r.GET("/api/timetable/:id", getTimetableHandler)
	r.PUT("/api/timetable/:id/approve", approveTimetableHandler)
	r.PUT("/api/timetable/:id/revert", revertTimetableHandler)
	r.PUT("/api/timetable/:id/slot/:slotId/move", moveCourseHandler)
	r.POST("/api/timetable/:id/slot/:slotId/check-conflicts", checkMoveConflictsHandler)
	r.DELETE("/api/timetable/:id", deleteTimetableHandler)

	// Dashboard stats
	r.GET("/api/dashboard/stats", getDashboardStats)

	// PDF export
	r.GET("/api/timetable/:id/pdf", getPDFHandler)
	r.GET("/api/timetable/:id/session/:day/:period/students", getSessionStudentsHandler)

	fmt.Println("🚀 Go Backend Server running on http://localhost:8080")
	r.Run(":8080")
}
