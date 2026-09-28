package database

import (
	"fmt"
	"google_login/models"
)

func Google_Response(data models.GoogleInformationResponse) error {

	data_response := map[string]interface{}{
		"status":  "success",
		"message": "Login successful",
		"data":    data,
	}

	fmt.Println(data_response)
	return nil
	// return data_response, nil
}
