package clash

import (
	"database/sql"
)

// GetStudentSet returns a set of register_nos for a given course_code and regulation_id
func GetStudentSet(db *sql.DB, courseCode string, regulationID int, onlyArrear bool) (map[string]bool, error) {
	query := `SELECT DISTINCT register_no FROM student_data 
		 WHERE course_code = ? AND regulation_id = ?`
	if onlyArrear {
		query += " AND upload_type = 'Arrear'"
	}
	rows, err := db.Query(query, courseCode, regulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	set := make(map[string]bool)
	for rows.Next() {
		var reg string
		if err := rows.Scan(&reg); err == nil {
			set[reg] = true
		}
	}
	return set, nil
}

// HasClash returns true if ANY student is enrolled in both courseA and courseB
func HasClash(setA, setB map[string]bool) bool {
	for reg := range setA {
		if setB[reg] {
			return true
		}
	}
	return false
}

// Intersection returns the list of students enrolled in both courses (for reporting)
func Intersection(setA, setB map[string]bool) []string {
	var clashing []string
	for reg := range setA {
		if setB[reg] {
			clashing = append(clashing, reg)
		}
	}
	return clashing
}
