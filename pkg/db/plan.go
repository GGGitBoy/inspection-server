package db

import (
	"database/sql"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func GetPlan(planID string) (*apis.Plan, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, timer, cron, mode, state, template_id, notify_id FROM plan WHERE id = ? LIMIT 1", planID)

	var id, name, timer, cron, state, templateID, notifyID string
	var mode int
	err = row.Scan(&id, &name, &timer, &cron, &mode, &state, &templateID, &notifyID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	}

	return &apis.Plan{
		ID:         id,
		Name:       name,
		Timer:      timer,
		Cron:       cron,
		Mode:       mode,
		State:      state,
		TemplateID: templateID,
		NotifyID:   notifyID,
	}, nil
}

func ListPlan() ([]*apis.Plan, error) {
	DB, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	rows, err := DB.Query("SELECT id, name, timer, cron, mode, state, template_id, notify_id FROM plan")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	plans := apis.NewPlans()
	for rows.Next() {
		var id, name, timer, cron, state, templateID, notifyID string
		var mode int
		err = rows.Scan(&id, &name, &timer, &cron, &mode, &state, &templateID, &notifyID)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("没有找到匹配的数据")
			} else {
				return nil, err
			}
		}

		plans = append(plans, &apis.Plan{
			ID:         id,
			Name:       name,
			Timer:      timer,
			Cron:       cron,
			Mode:       mode,
			State:      state,
			TemplateID: templateID,
			NotifyID:   notifyID,
		})
	}

	return plans, nil
}

func CreatePlan(plan *apis.Plan) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO plan(id, name, timer, cron, mode, state, template_id, notify_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(plan.ID, plan.Name, plan.Timer, plan.Cron, plan.Mode, plan.State, plan.TemplateID, plan.NotifyID)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func UpdatePlan(plan *apis.Plan) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("UPDATE plan SET name = ?, timer = ?, cron = ?, state = ?, template_id = ?, notify_id = ? WHERE id = ?", plan.Name, plan.Timer, plan.Cron, plan.State, plan.TemplateID, plan.NotifyID, plan.ID)
	if err != nil {
		return err
	}

	return nil
}

func DeletePlan(planID string) error {
	DB, err := GetDB()
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM plan WHERE id = ?", planID)
	if err != nil {
		return err
	}

	return nil
}
