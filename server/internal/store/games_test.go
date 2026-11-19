package store

import (
	"errors"
	"testing"
	"time"
)

func play(t *testing.T, s *Store, game string, score int64) GameResult {
	t.Helper()
	res, err := s.SaveGameScore(ctx, game, score)
	if err != nil {
		t.Fatalf("score %d in %s: %v", score, game, err)
	}
	return res
}

func shiftOf(t *testing.T, s *Store, room, name string) {
	t.Helper()
	if _, err := s.OpenShift(ctx, room, name, 0); err != nil {
		t.Fatal(err)
	}
}

func endShift(t *testing.T, s *Store) {
	t.Helper()
	if _, err := s.db.ExecContext(ctx, "UPDATE shifts SET closed_at = 1 WHERE closed_at IS NULL"); err != nil {
		t.Fatal(err)
	}
}

func TestGameScoreNeedsAShiftAndAKnownGame(t *testing.T) {
	s := newStore(t)
	if _, err := s.SaveGameScore(ctx, "flyer", 5); !errors.Is(err, ErrNoOpenShift) {
		t.Errorf("without a shift: %v", err)
	}
	open(t, s, 0)
	for _, c := range []struct {
		game  string
		score int64
	}{{"tetris", 5}, {"snake", 5}, {"", 5}, {"flyer", -1}, {"flyer", maxGameScore + 1}} {
		if _, err := s.SaveGameScore(ctx, c.game, c.score); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q %d: %v", c.game, c.score, err)
		}
	}
}

func TestGameBoardKeepsEachWorkersBestRun(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 14, 21, 0, 0, 0, time.Local)
	at(s, base)
	shiftOf(t, s, "214", "Jonas")
	if res := play(t, s, "flyer", 12); !res.Saved || !res.NewBest || res.Best != 12 || *res.Place != 1 || res.Record != 12 {
		t.Errorf("first run = %+v", res)
	}
	at(s, base.Add(time.Minute))
	if res := play(t, s, "flyer", 7); !res.Saved || res.NewBest || res.Best != 12 || *res.Place != 1 {
		t.Errorf("a worse run = %+v", res)
	}
	endShift(t, s)

	at(s, base.Add(time.Hour))
	shiftOf(t, s, "112", "Lena")
	if res := play(t, s, "flyer", 12); res.NewBest != true || *res.Place != 2 || res.Record != 12 {
		t.Errorf("the same score later must stay behind = %+v", res)
	}
	at(s, base.Add(2*time.Hour))
	if res := play(t, s, "flyer", 30); *res.Place != 1 || res.Record != 30 || len(res.Board.Top) != 2 {
		t.Errorf("a new record = %+v", res)
	}
	endShift(t, s)

	// The same person with other capitals is still one row.
	at(s, base.Add(3*time.Hour))
	shiftOf(t, s, "214", "jonas")
	if res := play(t, s, "flyer", 20); !res.NewBest || *res.Place != 2 || len(res.Board.Top) != 2 {
		t.Errorf("Jonas again = %+v", res)
	}

	boards, err := s.GameBoards(ctx)
	if err != nil {
		t.Fatal(err)
	}
	top := boards["flyer"].Top
	if len(top) != 2 || top[0].Name != "Lena" || top[0].Score != 30 || top[1].Room != "214" || top[1].Score != 20 {
		t.Errorf("flyer board = %+v", top)
	}
	if top[1].At != base.Add(3*time.Hour).UnixMilli() {
		t.Errorf("the date is the best run's, got %d", top[1].At)
	}
	if mine := boards["flyer"].Mine; mine == nil || mine.Place != 2 || mine.Score != 20 {
		t.Errorf("mine = %+v", mine)
	}
}

func TestGameRunOfNothingIsNotStored(t *testing.T) {
	s := newStore(t)
	open(t, s, 0)
	res := play(t, s, "flyer", 0)
	if res.Saved || res.NewBest || res.Place != nil || res.Board.Top == nil || len(res.Board.Top) != 0 {
		t.Errorf("a run of 0 = %+v", res)
	}
	boards, err := s.GameBoards(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if b, ok := boards["flyer"]; !ok || b.Top == nil || len(b.Top) != 0 || b.Mine != nil {
		t.Errorf("an empty board must still be a list: %+v", b)
	}
}

func TestGameBoardShowsTenAndStillFindsMe(t *testing.T) {
	s := newStore(t)
	for i := range 12 {
		shiftOf(t, s, "1"+string(rune('a'+i)), "Gast")
		play(t, s, "flyer", int64(100-i))
		endShift(t, s)
	}
	shiftOf(t, s, "999", "Letzter")
	res := play(t, s, "flyer", 1)
	if len(res.Board.Top) != boardSize || *res.Place != 13 || res.Board.Mine == nil || res.Board.Mine.Place != 13 {
		t.Errorf("place below the board = %+v, top %d", res.Place, len(res.Board.Top))
	}
	endShift(t, s)
	boards, err := s.GameBoards(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if boards["flyer"].Mine != nil {
		t.Error("with nobody on shift there is no own row")
	}
}

func TestUpgradeFromVersionSixAddsTheGames(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	asVersionEight(t, s)
	for _, q := range []string{"DROP TABLE game_scores", "PRAGMA user_version = 6"} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	s.Close()
	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	open(t, s, 0)
	if res := play(t, s, "flyer", 4); !res.Saved {
		t.Errorf("after the upgrade = %+v", res)
	}
}
