// AMK Kasse: the till for the bar at the Albertus Magnus dorm in Münster.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"amkpos/internal/api"
	"amkpos/internal/store"
	"amkpos/internal/webui"
)

func main() {
	dataDir := flag.String("data", "", "folder for the database and backups (default: \"data\" next to the program)")
	port := flag.Int("port", 8417, "port to listen on")
	lan := flag.Bool("lan", false, "allow other devices on the network to reach the admin area")
	noBrowser := flag.Bool("no-browser", false, "do not open the till window")
	windowed := flag.Bool("windowed", false, "open a normal window instead of full-screen kiosk mode")
	demo := flag.Bool("demo", false, "start with invented sales in a separate database")
	resetPIN := flag.Bool("reset-pin", false, "clear the admin PIN, then exit")
	flag.Parse()

	// A soft memory ceiling keeps the program small on a weak PC.
	debug.SetMemoryLimit(48 << 20)
	debug.SetGCPercent(50)

	dir := resolveDataDir(*dataDir, *demo)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal("Der Datenordner konnte nicht angelegt werden: %v", err)
	}
	logFile, err := os.OpenFile(filepath.Join(dir, "amk-kasse.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	}

	// Works while the till is running, because SQLite allows a second connection.
	if *resetPIN {
		st, err := store.Open(dir)
		if err != nil {
			fatal("Die Datenbank konnte nicht geöffnet werden: %v", err)
		}
		defer st.Close()
		if err := st.DeleteSetting(context.Background(), api.KeyPIN); err != nil {
			fatal("Die PIN konnte nicht zurückgesetzt werden: %v", err)
		}
		showMessage("Die OBK-PIN wurde gelöscht. Beim nächsten Öffnen des OBK-Bereichs wird eine neue festgelegt.")
		return
	}

	host := "127.0.0.1"
	if *lan {
		host = "0.0.0.0"
	}
	addr := net.JoinHostPort(host, fmt.Sprint(*port))
	url := fmt.Sprintf("http://127.0.0.1:%d/", *port)

	// A second start just shows the window of the copy already running.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		if alreadyRunning(url) {
			if !*noBrowser {
				openBrowser(url, dir, *windowed)
			}
			return
		}
		fatal("Port %d wird schon von einem anderen Programm benutzt: %v", *port, err)
	}

	// Only after the port check, so a second demo cannot wipe a running one.
	if *demo {
		if err := resetDemoData(dir); err != nil {
			fatal("Die Demo-Daten konnten nicht zurückgesetzt werden: %v", err)
		}
	}

	st, err := store.Open(dir)
	if err != nil {
		fatal("Die Datenbank konnte nicht geöffnet werden: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	if *demo {
		if err := st.SeedDemoHistory(ctx); err != nil {
			fatal("Die Demo-Daten konnten nicht angelegt werden: %v", err)
		}
	}

	app := api.New(st, webui.FS())
	app.Demo = *demo
	server := &http.Server{
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	if !*demo {
		go app.DailyBackup(ctx)
	}
	go app.OnlineSync(ctx)

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server: %v", err)
			stop()
		}
	}()
	log.Printf("AMK Kasse is running at %s (data: %s)", url, dir)
	if !*noBrowser {
		openBrowser(url, dir, *windowed)
	}

	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// fatal has to show a window: the program has no console, so a log line alone
// would look like a double click that did nothing.
func fatal(format string, args ...any) {
	log.Printf(format, args...)
	showMessage(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func resetDemoData(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "amk-bar.db*"))
	if err != nil {
		return err
	}
	for _, file := range append(files, filepath.Join(dir, "backups")) {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}
	return nil
}

func resolveDataDir(flagValue string, demo bool) string {
	if flagValue != "" {
		return flagValue
	}
	base := "."
	if exe, err := os.Executable(); err == nil {
		base = filepath.Dir(exe)
	}
	if demo {
		return filepath.Join(base, "data-demo")
	}
	return filepath.Join(base, "data")
}

func alreadyRunning(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// openBrowser shows the till in a dedicated browser profile, which guarantees a
// real kiosk window even if the browser is already open for something else.
func openBrowser(url, dataDir string, windowed bool) {
	profile := filepath.Join(dataDir, "browser-profile")
	mode := []string{"--kiosk", url, "--edge-kiosk-type=fullscreen"}
	chromeMode := []string{"--kiosk", url}
	if windowed {
		mode = []string{"--app=" + url}
		chromeMode = mode
	}
	common := []string{"--no-first-run", "--no-default-browser-check", "--user-data-dir=" + profile,
		// The browser is the heaviest part by far, so everything a till does not need stays off.
		"--renderer-process-limit=1", "--disable-extensions", "--disable-component-extensions-with-background-pages", "--disable-sync", "--disable-default-apps",
		"--disable-background-networking", "--disable-component-update", "--disable-domain-reliability",
		"--disable-features=Translate,msEdgeSidebarV2,msEdgeCollections,msShoppingExp,EdgeWorkspacesEnabled"}

	for _, candidate := range browserCandidates() {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		args := chromeMode
		if filepath.Base(candidate) == "msedge.exe" {
			args = mode
		}
		if err := exec.Command(path, append(common, args...)...).Start(); err == nil {
			return
		}
	}

	// No known browser: fall back to whatever the system opens links with.
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		log.Printf("could not open a browser: %v", err)
	}
}

func browserCandidates() []string {
	if runtime.GOOS == "windows" {
		return []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
	}
	return []string{"chromium", "chromium-browser", "google-chrome"}
}
