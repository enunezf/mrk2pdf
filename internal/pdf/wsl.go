package pdf

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// newWSLAllocator manually launches a Windows browser binary from WSL with
// the DevTools endpoint bound to all interfaces, then rewrites the reported
// localhost WS URL to the Windows host IP so chromedp (running inside WSL2's
// NAT'd network namespace) can actually reach it.
//
// This is necessary because chromedp.NewExecAllocator captures the URL chrome
// prints to stderr and dials it verbatim. Windows chrome reports 127.0.0.1
// (its own loopback), which is unreachable from WSL2 unless the user has
// enabled mirrored networking (only available on recent Windows 11 builds).
func newWSLAllocator(parent context.Context, browserPath string) (*allocator, error) {
	winUDD := fmt.Sprintf(`C:\Windows\Temp\mrk2pdf-%d-%d`, os.Getpid(), time.Now().UnixNano())
	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-background-networking",
		"--disable-extensions",
		"--disable-dev-shm-usage",
		"--hide-scrollbars",
		"--mute-audio",
		"--remote-debugging-port=0",
		"--remote-debugging-address=0.0.0.0",
		"--user-data-dir=" + winUDD,
		"about:blank",
	}
	cmd := exec.Command(browserPath, args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("creando pipe stderr: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("lanzando %s: %w", browserPath, err)
	}

	wsURL, err := readDevToolsURL(stderr, 20*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, err
	}
	go io.Copy(io.Discard, stderr)

	hostIP, err := windowsHostIP()
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("determinando IP del host Windows: %w", err)
	}
	wsURL = rewriteLocalhost(wsURL, hostIP)

	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(parent, wsURL)
	cancel := func() {
		cancelAlloc()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	return &allocator{ctx: allocCtx, cancel: cancel}, nil
}

// readDevToolsURL scans chrome's stderr for the standard
// "DevTools listening on ws://..." line and returns the URL.
func readDevToolsURL(stderr io.Reader, timeout time.Duration) (string, error) {
	const prefix = "DevTools listening on "
	type result struct {
		url string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if i := strings.Index(line, prefix); i >= 0 {
				ch <- result{url: strings.TrimSpace(line[i+len(prefix):])}
				return
			}
		}
		ch <- result{err: fmt.Errorf("stderr terminó sin URL de DevTools: %v", scanner.Err())}
	}()
	select {
	case r := <-ch:
		return r.url, r.err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout esperando URL de DevTools")
	}
}

// windowsHostIP returns the IP address WSL2 uses to reach the Windows host,
// derived from the default route gateway.
func windowsHostIP() (string, error) {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", fmt.Errorf("ip route show default: %w", err)
	}
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "via" && i+1 < len(fields) {
			return fields[i+1], nil
		}
	}
	return "", fmt.Errorf("no se encontró 'via <gateway>' en la salida de ip route")
}

// rewriteLocalhost replaces a loopback host inside a WS URL with the given IP,
// so chromedp can reach a chrome instance running outside its network namespace.
func rewriteLocalhost(wsURL, hostIP string) string {
	for _, host := range []string{"127.0.0.1", "0.0.0.0", "localhost"} {
		needle := "://" + host + ":"
		if strings.Contains(wsURL, needle) {
			return strings.Replace(wsURL, needle, "://"+hostIP+":", 1)
		}
	}
	return wsURL
}
