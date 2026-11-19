package api

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"amkpos/internal/store"
)

// money formats cents the way German Excel reads numbers: 1234,50.
func money(cents int64) string {
	sign := ""
	if cents < 0 {
		sign, cents = "-", -cents
	}
	return fmt.Sprintf("%s%d,%02d", sign, cents/100, cents%100)
}

func clock(ms int64) string {
	return time.UnixMilli(ms).Local().Format("02.01.2006 15:04")
}

// safeText stops spreadsheets from running a name that starts like a formula.
func safeText(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func paymentName(p string) string {
	if p == store.PayCash {
		return "Bar"
	}
	return "Unbar"
}

func optional(v *int64) string {
	if v == nil {
		return ""
	}
	return money(*v)
}

// export streams a CSV that German Excel opens correctly: semicolons plus a BOM.
func (a *Server) export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var header []string
	var rows [][]string
	var name string

	switch r.PathValue("kind") {
	case "sales.csv":
		from, to, valid := dateRange(w, r)
		if !valid {
			return
		}
		sales, err := a.store.SalesBetween(ctx, from, to, 1_000_000)
		if err != nil {
			fail(w, err)
			return
		}
		name = "verkaeufe-" + from + "-bis-" + to
		header = []string{"Geschäftstag", "Zeit", "Beleg", "Schicht", "Artikel", "Kategorie", "Menge", "Einzelpreis", "Summe", "Zahlart", "Storno von", "Grund"}
		for _, s := range sales {
			shift, voids := "", ""
			if s.ShiftID != nil {
				shift = strconv.FormatInt(*s.ShiftID, 10)
			}
			if s.VoidsSaleID != nil {
				voids = strconv.FormatInt(*s.VoidsSaleID, 10)
			}
			for _, l := range s.Lines {
				rows = append(rows, []string{s.Day, clock(s.CreatedAt), strconv.FormatInt(s.ID, 10), shift, safeText(l.Name),
					safeText(l.Category), strconv.FormatInt(l.Qty, 10), money(l.UnitPriceCents), money(l.Qty * l.UnitPriceCents),
					paymentName(s.Payment), voids, safeText(s.Reason)})
			}
		}
	case "products.csv":
		from, to, valid := dateRange(w, r)
		if !valid {
			return
		}
		rep, err := a.store.Report(ctx, from, to)
		if err != nil {
			fail(w, err)
			return
		}
		name = "artikel-" + from + "-bis-" + to
		header = []string{"Artikel", "Kategorie", "Menge", "Umsatz", "Einkauf", "Marge"}
		for _, p := range rep.Products {
			rows = append(rows, []string{safeText(p.Name), safeText(p.Category), strconv.FormatInt(p.Qty, 10),
				money(p.RevenueCents), money(p.CostCents), money(p.RevenueCents - p.CostCents)})
		}
	case "shifts.csv":
		shifts, err := a.store.ListShifts(ctx, 100_000)
		if err != nil {
			fail(w, err)
			return
		}
		name = "schichten"
		header = []string{"Schicht", "Zimmer", "Name", "Beginn", "Ende", "Anfangsbestand", "Umsatz", "Soll", "Gezählt", "Differenz"}
		for _, s := range shifts {
			end, diff := "", ""
			if s.ClosedAt != nil {
				end = clock(*s.ClosedAt)
			}
			if s.ExpectedCents != nil && s.CountedCents != nil {
				diff = money(*s.CountedCents - *s.ExpectedCents)
			}
			rows = append(rows, []string{strconv.FormatInt(s.ID, 10), safeText(s.WorkerRoom), safeText(s.WorkerName),
				clock(s.OpenedAt), end, money(s.FloatCents), money(s.TotalCents), optional(s.ExpectedCents), optional(s.CountedCents), diff})
		}
	case "stock.csv":
		products, err := a.store.Products(ctx)
		if err != nil {
			fail(w, err)
			return
		}
		name = "bestand-" + a.store.Now().Format("2006-01-02")
		header = []string{"Artikel", "An der Bar", "Im Lager", "Gesamt", "Soll an der Bar", "Mindestbestand", "Platz an der Bar", "Einkaufspreis", "Warenwert"}
		for _, p := range products {
			if p.TrackStock && p.Active {
				total := p.BarStock + p.StorageStock
				rows = append(rows, []string{safeText(p.Name), strconv.FormatInt(p.BarStock, 10), strconv.FormatInt(p.StorageStock, 10),
					strconv.FormatInt(total, 10), strconv.FormatInt(p.BarTarget, 10), strconv.FormatInt(p.MinStock, 10),
					safeText(p.Location), money(p.CostCents), money(max(total, 0) * p.CostCents)})
			}
		}
	default:
		fail(w, store.ErrNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, name))
	if _, err := w.Write([]byte("\xEF\xBB\xBF")); err != nil {
		return
	}
	out := csv.NewWriter(w)
	out.Comma = ';'
	out.UseCRLF = true
	out.Write(header)
	out.WriteAll(rows)
	if err := out.Error(); err != nil {
		log.Printf("export: %v", err)
	}
}
