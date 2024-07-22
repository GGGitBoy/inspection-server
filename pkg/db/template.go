package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func GetTemplate(templateID string) (*apis.Template, error) {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, data FROM template WHERE id = ? LIMIT 1", templateID)

	var id, name, data string
	template := apis.NewTemplate()
	err = row.Scan(&id, &name, &data)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	} else {
		var dataKubernetesConfig []*apis.KubernetesConfig
		err := json.Unmarshal([]byte(data), &dataKubernetesConfig)
		if err != nil {
			return nil, err
		}

		template = &apis.Template{
			ID:               id,
			Name:             name,
			KubernetesConfig: dataKubernetesConfig,
		}
	}

	return template, nil
}

func ListTemplate() ([]*apis.Template, error) {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	rows, err := DB.Query("SELECT id, name, data FROM template")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	templates := apis.NewTemplates()
	for rows.Next() {
		var id, name, data string
		err = rows.Scan(&id, &name, &data)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("没有找到匹配的数据")
			} else {
				return nil, err
			}
		}

		var dataKubernetesConfig []*apis.KubernetesConfig
		err := json.Unmarshal([]byte(data), &dataKubernetesConfig)
		if err != nil {
			return nil, err
		}

		templates = append(templates, &apis.Template{
			ID:               id,
			Name:             name,
			KubernetesConfig: dataKubernetesConfig,
		})
	}

	return templates, nil
}

func CreateTemplate(template *apis.Template) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	data, err := json.Marshal(template.KubernetesConfig)
	if err != nil {
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO template(id, name, data) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(template.ID, template.Name, string(data))
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func UpdateTemplate(template *apis.Template) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return err
	}
	defer DB.Close()

	data, err := json.Marshal(template.KubernetesConfig)
	if err != nil {
		return err
	}

	_, err = DB.Exec("UPDATE template SET name = ?, data = ? WHERE id = ?", template.Name, string(data), template.ID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTemplate(templateID string) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM template WHERE id = ?", templateID)
	if err != nil {
		return err
	}

	return nil
}
