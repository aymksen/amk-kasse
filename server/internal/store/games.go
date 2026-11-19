package store

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
)

// The mini-games behind the logo keep a leaderboard each. A score belongs to whoever is on shift, so room and name
// are copied from the shift, the way a receipt keeps the worker. Scores stay on the till: they are not money, stock
// or sales, and they never go to the online dashboard.

// gamesSQL is schema version 7.
const gamesSQL = `
CREATE TABLE IF NOT EXISTS game_scores (
  id         INTEGER PRIMARY KEY,
  game       TEXT NOT NULL,
  shift_id   INTEGER NOT NULL REFERENCES shifts(id),
  room       TEXT NOT NULL,
  name       TEXT NOT NULL,
  score      INTEGER NOT NULL CHECK (score > 0),
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS game_scores_by_game ON game_scores(game, score DESC);
`

// GameIDs are the games the till knows. The web app names them the same way, and a score for any other name is
// refused.
var GameIDs = []string{"flyer"}

const (
	boardSize = 10
	// No run gets anywhere near this. It only keeps a made-up number off the board.
	maxGameScore = 1_000_000
)

// GameScore is one worker's best run in a game.
type GameScore struct {
	Name  string `json:"name"`
	Room  string `json:"room"`
	Score int64  `json:"score"`
	At    int64  `json:"at"`
}

// MyScore is the best run of the worker on shift, with their place, which can be below the board.
type MyScore struct {
	GameScore
	Place int `json:"place"`
}

// GameBoard is a game's top ten, plus the worker on shift.
type GameBoard struct {
	Top  []GameScore `json:"top"`
	Mine *MyScore    `json:"mine"`
}

// GameResult is what a game hears back after a run.
type GameResult struct {
	Saved   bool      `json:"saved"`
	NewBest bool      `json:"newBest"`
	Best    int64     `json:"best"`
	Place   *int      `json:"place"`
	Record  int64     `json:"record"`
	Board   GameBoard `json:"board"`
}

type player struct{ room, name string }

// key tells workers apart the way a person would: the same room and the same name, whatever the capitals.
func (p player) key() string { return p.room + "\x00" + strings.ToLower(p.name) }

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// gameBoard ranks each worker by their best run. At a tie, whoever got there first stays ahead.
func gameBoard(ctx context.Context, q queryer, game string, me *player) (GameBoard, error) {
	rows, err := q.QueryContext(ctx, `SELECT room, name, score, created_at FROM game_scores
		WHERE game = ? ORDER BY score DESC, created_at, id`, game)
	if err != nil {
		return GameBoard{}, err
	}
	defer rows.Close()
	board := GameBoard{Top: []GameScore{}}
	seen := map[string]bool{}
	place := 0
	for rows.Next() {
		var row GameScore
		if err := rows.Scan(&row.Room, &row.Name, &row.Score, &row.At); err != nil {
			return GameBoard{}, err
		}
		who := player{row.Room, row.Name}
		if seen[who.key()] {
			continue
		}
		seen[who.key()] = true
		place++
		if place <= boardSize {
			board.Top = append(board.Top, row)
		}
		if me != nil && who.key() == me.key() {
			board.Mine = &MyScore{row, place}
		}
	}
	return board, rows.Err()
}

func onShift(ctx context.Context, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}) (int64, *player, error) {
	var id int64
	var p player
	err := q.QueryRowContext(ctx, "SELECT id, worker_room, worker_name FROM shifts WHERE closed_at IS NULL").
		Scan(&id, &p.room, &p.name)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil, nil
	}
	return id, &p, err
}

// GameBoards returns every game's board, with the worker on shift picked out if there is one.
func (s *Store) GameBoards(ctx context.Context) (map[string]GameBoard, error) {
	_, me, err := onShift(ctx, s.db)
	if err != nil {
		return nil, err
	}
	boards := map[string]GameBoard{}
	for _, game := range GameIDs {
		if boards[game], err = gameBoard(ctx, s.db, game, me); err != nil {
			return nil, err
		}
	}
	return boards, nil
}

// SaveGameScore stores a finished run for the worker on shift. A run of 0 is not stored but still answered, so the
// game can show the board as it is.
func (s *Store) SaveGameScore(ctx context.Context, game string, score int64) (GameResult, error) {
	var res GameResult
	if !slices.Contains(GameIDs, game) || score < 0 || score > maxGameScore {
		return res, ErrInvalid
	}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		shiftID, me, err := onShift(ctx, tx)
		if err != nil {
			return err
		}
		if me == nil {
			return ErrNoOpenShift
		}
		before, err := gameBoard(ctx, tx, game, me)
		if err != nil {
			return err
		}
		if score > 0 {
			if _, err := tx.ExecContext(ctx, `INSERT INTO game_scores(game, shift_id, room, name, score, created_at)
				VALUES (?, ?, ?, ?, ?, ?)`, game, shiftID, me.room, me.name, score, s.now().UnixMilli()); err != nil {
				return err
			}
		}
		after, err := gameBoard(ctx, tx, game, me)
		if err != nil {
			return err
		}
		res = GameResult{Saved: score > 0, Board: after}
		if after.Mine != nil {
			res.Best = after.Mine.Score
			if score > 0 {
				place := after.Mine.Place
				res.Place = &place
			}
		}
		res.NewBest = score > 0 && (before.Mine == nil || score > before.Mine.Score)
		if len(after.Top) > 0 {
			res.Record = after.Top[0].Score
		}
		return nil
	})
	return res, err
}
