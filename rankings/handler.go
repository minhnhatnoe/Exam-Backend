package rankings

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/The-Codefun-Exam-Team/Exam-Backend/general"
	"github.com/The-Codefun-Exam-Team/Exam-Backend/models"
)

type JSONRanking struct {
	Avatar    string  `json:"avatar"`
	Gid       int     `json:"group"`
	Groupname string  `json:"groupname"`
	Tid       int     `json:"id"`
	Teamname  string  `json:"username"`
	Name      string  `json:"name"`
	Rank      int     `json:"rank"`
	Score     float64 `json:"points"`
}

func (g *Group) RankingsGet(c echo.Context) error {
	var ranking []JSONRanking

	var rows *sql.Rows
	var err error

	models.ResolveQueue(g.db)

	str_groupid := c.FormValue("group")
	groupid, err := strconv.Atoi(str_groupid)
	if err != nil {
		return err
	}

	str_pageid := c.FormValue("pageid")
	pageid, err := strconv.Atoi(str_pageid)
	if err != nil {
		return err
	}

	str_limit := c.FormValue("limit")
	limit, err := strconv.Atoi(str_limit)
	if err != nil {
		return err
	}

	log.Print("Converted input")

	if groupid > 0 {
		rows, err = g.db.Query(`WITH score_table AS (SELECT tid, dpid, MAX(score) AS score FROM debug_submissions GROUP BY tid, dpid)
		SELECT teams.email, groups.gid, groups.groupname, teams.tid, teams.teamname, teams.name, SUM(score_table.score) AS score FROM score_table
		INNER JOIN teams ON teams.tid = score_table.tid
		INNER JOIN groups ON groups.gid = teams.group
		WHERE teams.group = ?
		GROUP BY score_table.tid ORDER BY score DESC
		LIMIT ? OFFSET ?`, groupid, limit, (pageid-1)*limit)
	} else {
		rows, err = g.db.Query(`WITH score_table AS (SELECT tid, dpid, MAX(score) AS score FROM debug_submissions GROUP BY tid, dpid)
		SELECT teams.email, groups.gid, groups.groupname, teams.tid, teams.teamname, teams.name, SUM(score_table.score) AS score FROM score_table 
		INNER JOIN teams ON teams.tid = score_table.tid
		INNER JOIN groups ON groups.gid = teams.group
		GROUP BY score_table.tid ORDER BY score DESC
		LIMIT ? OFFSET ?`, limit, (pageid-1)*limit)
	}

	log.Print("Query from db")

	defer rows.Close()

	if err != nil {
		return err
	}

	rank := (pageid-1)*limit + 1
	for rows.Next() {
		var email, teamname, name, groupname string
		var gid, tid int
		var score float64

		if err := rows.Scan(&email, &gid, &groupname, &tid, &teamname, &name, &score); err != nil {
			return err
		}

		ranking = append(ranking, JSONRanking{
			Avatar:    general.GetHash(email),
			Gid:       gid,
			Groupname: groupname,
			Tid:       tid,
			Teamname:  teamname,
			Name:      name,
			Rank:      rank,
			Score:     score,
		})
		rank++
	}

	log.Print("Iterated")

	if err := rows.Err(); err != nil {
		return err
	}

	log.Print("Checked for scan errors")

	return c.JSON(http.StatusOK, ranking)
}
