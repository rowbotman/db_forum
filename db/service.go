package db

import "github.com/rowbotman/db_forum/models"

func ServiceGet() (models.ServiceInfo, error) {
	sqlStatement := `SELECT COUNT(*) FROM profile`
	row := DB.QueryRow(sqlStatement)
	info := models.ServiceInfo{}
	if err := row.Scan(&info.User); err != nil {
		return models.ServiceInfo{}, err
	}
	sqlStatement = `SELECT COUNT(*) FROM forum`
	row = DB.QueryRow(sqlStatement)
	if err := row.Scan(&info.Forum); err != nil {
		return models.ServiceInfo{}, err
	}
	sqlStatement = `SELECT COUNT(*) FROM thread`
	row = DB.QueryRow(sqlStatement)
	if err := row.Scan(&info.Thread); err != nil {
		return models.ServiceInfo{}, err
	}
	sqlStatement = `SELECT COUNT(*) FROM post`
	row = DB.QueryRow(sqlStatement)
	if err := row.Scan(&info.Post); err != nil {
		return models.ServiceInfo{}, err
	}
	return info, nil
}


func ClearService() bool {
	sqlStatement := `TRUNCATE TABLE profile, forum, thread, post, vote, forum_meta CASCADE;`
	if _, err := DB.Exec(sqlStatement); err != nil {
		return false
	}
	return true
}