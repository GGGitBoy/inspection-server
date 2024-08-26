package db

import (
	"encoding/json"
	"inspection-server/pkg/apis"
	"log"
)

// GetTemplate retrieves a template from the database by templateID.
func GetTemplate(templateID string) (*apis.Template, error) {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return nil, err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	row := DB.QueryRow("SELECT id, name, data FROM template WHERE id = ? LIMIT 1", templateID)

	var id, name, data string
	template := apis.NewTemplate()
	err = row.Scan(&id, &name, &data)
	if err != nil {
		log.Printf("Error scanning template row: %v", err)
		return nil, err
	}

	var dataKubernetesConfig []*apis.KubernetesConfig
	err = json.Unmarshal([]byte(data), &dataKubernetesConfig)
	if err != nil {
		log.Printf("Error unmarshalling KubernetesConfig: %v", err)
		return nil, err
	}

	template = &apis.Template{
		ID:               id,
		Name:             name,
		KubernetesConfig: dataKubernetesConfig,
	}

	log.Printf("Template retrieved successfully with ID: %s", template.ID)
	return template, nil
}

// ListTemplate retrieves all templates from the database.
func ListTemplate() ([]*apis.Template, error) {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return nil, err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	rows, err := DB.Query("SELECT id, name, data FROM template")
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	templates := apis.NewTemplates()
	for rows.Next() {
		var id, name, data string
		err = rows.Scan(&id, &name, &data)
		if err != nil {
			log.Printf("Error scanning template row: %v", err)
			return nil, err
		}

		var dataKubernetesConfig []*apis.KubernetesConfig
		err = json.Unmarshal([]byte(data), &dataKubernetesConfig)
		if err != nil {
			log.Printf("Error unmarshalling KubernetesConfig: %v", err)
			return nil, err
		}

		templates = append(templates, &apis.Template{
			ID:               id,
			Name:             name,
			KubernetesConfig: dataKubernetesConfig,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over template rows: %v", err)
		return nil, err
	}

	log.Printf("Templates retrieved successfully, total count: %d", len(templates))
	return templates, nil
}

// CreateTemplate inserts a new template into the database.
func CreateTemplate(template *apis.Template) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	data, err := json.Marshal(template.KubernetesConfig)
	if err != nil {
		log.Printf("Error marshalling KubernetesConfig: %v", err)
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO template(id, name, data) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(template.ID, template.Name, string(data))
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Template created successfully with ID: %s", template.ID)
	return nil
}

// UpdateTemplate updates an existing template in the database.
func UpdateTemplate(template *apis.Template) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	data, err := json.Marshal(template.KubernetesConfig)
	if err != nil {
		log.Printf("Error marshalling KubernetesConfig: %v", err)
		return err
	}

	_, err = DB.Exec("UPDATE template SET name = ?, data = ? WHERE id = ?", template.Name, string(data), template.ID)
	if err != nil {
		log.Printf("Error updating template: %v", err)
		return err
	}

	log.Printf("Template updated successfully with ID: %s", template.ID)
	return nil
}

// DeleteTemplate removes a template from the database by templateID.
func DeleteTemplate(templateID string) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer func() {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	result, err := DB.Exec("DELETE FROM template WHERE id = ?", templateID)
	if err != nil {
		log.Printf("Error executing delete statement: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No template found to delete with ID: %s", templateID)
	} else {
		log.Printf("Template deleted successfully with ID: %s", templateID)
	}

	return nil
}
