package database

import (
	"database/sql"
	"fmt"
	"google_login/models"
)

// var DB *sql.DB

// ======= Create ======

func SaveToDB(u models.GoogleInformationResponse) error {
	var birthDate sql.NullString
	var pid sql.NullString
	if u.BirthDate == "" || u.PID == "" {
		birthDate = sql.NullString{Valid: false} // จะกลายเป็น NULL ใน Database
		pid = sql.NullString{Valid: false}
	} else {
		birthDate = sql.NullString{String: u.BirthDate, Valid: true}
		pid = sql.NullString{String: u.PID, Valid: true}
	}

	query := `INSERT INTO users(pid, given_name, family_name, birthdate) VALUES ($1, $2, $3, $4)`
	_, err := DB.Exec(query, pid, u.GivenName, u.FamilyName, birthDate)
	if err == nil {
		fmt.Println("Save user to Database successfully")
		Google_Response(models.GoogleInformationResponse{
			PID:        u.PID,
			GivenName:  u.GivenName,
			FamilyName: u.FamilyName,
			BirthDate:  u.BirthDate,
		})
	} else {
		fmt.Println("Cannot save user")
	}
	return err
}
