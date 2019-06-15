package db

import (
	"errors"
	"github.com/jackc/pgx"
	"github.com/rowbotman/db_forum/models"
)


func InsertIntoUser(userData models.User) (models.Users, error) {
	var users models.Users
	sqlStatement := `SELECT full_name, nickname, email, about FROM profile WHERE LOWER(nickname) = LOWER($1) OR LOWER(email) = LOWER($2);`
	rows, err := DB.Query(sqlStatement, userData.Nickname, userData.Email)
	if err != nil && !rows.Next() {
		sqlStatement = `INSERT INTO profile VALUES (default, $1, $2, $3, $4);`

		_, err = DB.Exec(sqlStatement, userData.Nickname, userData.Name, userData.About, userData.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, userData)
		return users, nil
	}

	for rows.Next() {
		newUser := models.User{}
		err = rows.Scan(
			&newUser.Name,
			&newUser.Nickname,
			&newUser.Email,
			&newUser.About)
		if err != nil {
			// handle this error
			// but i did't know how to do this .-.
			return nil, err
		}
		users = append(users, newUser)
	}

	if len(users) == 0 {
		sqlStatement = `INSERT INTO profile VALUES (default, $1, $2, $3, $4);`

		_, err = DB.Exec(sqlStatement, userData.Nickname, userData.Name, userData.About, userData.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, userData)
		return users, nil
	}

	return users, errors.New("multiple rows")
}

func SelectUser(nickname string) (models.User, error) {
	sqlStatement := `SELECT uid, full_name, nickname, email, about FROM profile WHERE nickname = $1`
	row := DB.QueryRow(sqlStatement, nickname)
	newUser := models.User{}
	err := row.Scan(
		&newUser.Pk,
		&newUser.Name,
		&newUser.Nickname,
		&newUser.Email,
		&newUser.About)
	if err == pgx.ErrNoRows {
		return models.User{}, errors.New("no rows")
	} else if err != nil {
		return models.User{}, err
	}

	return newUser, nil
}

func UpdateUser(updUser models.User) (models.User, error) {
	sqlStatement := `
  SELECT full_name, nickname, email FROM profile WHERE LOWER(email) = LOWER($1);`
	row := DB.QueryRow(sqlStatement, updUser.Email)
	user := models.User{}
	err := row.Scan(
		&user.Name,
		&user.Nickname,
		&user.Email)
	if err == pgx.ErrNoRows || user.Nickname == updUser.Nickname {
		if updUser.IsEmpty() {
			userInfo, err := SelectUser(updUser.Nickname)
			if err != nil {
				return models.User{}, err
			}
			return userInfo, nil
		}

		userInfo, err := SelectUser(updUser.Nickname)
		if err != nil {
			return models.User{}, err
		}
		if len(updUser.About) == 0 {
			updUser.About = userInfo.About
		}
		if len(updUser.Name) == 0 {
			updUser.Name = userInfo.Name
		}
		if len(updUser.Email) == 0 {
			updUser.Email = userInfo.Email
		}

		sqlStatement = `UPDATE profile SET full_name = $1, email = $2, about = $3 WHERE nickname = $4;`
		_, err = DB.Exec(sqlStatement, updUser.Name, updUser.Email, updUser.About, updUser.Nickname)
		if err != nil {
			return models.User{}, err
		}

		return updUser, nil
	}

	return updUser, errors.New("This email is already registered by user: " + user.Nickname)
}