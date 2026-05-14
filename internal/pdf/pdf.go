package pdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	// A4 paper size expressed in inches, the unit expected by Chrome DevTools.
	a4WidthInches  = 8.27  // 210 mm
	a4HeightInches = 11.69 // 297 mm

	overallTimeout = 60 * time.Second
	mermaidTimeout = 15 * time.Second
)

// GeneratePDF launches a Chromium-based browser in headless mode, loads
// htmlContent via a data: URL, waits for any mermaid diagrams to finish
// rendering, and writes the resulting PDF to outputPath.
func GeneratePDF(htmlContent []byte, outputPath, browserPath string) error {
	alloc, err := newAllocator(context.Background(), browserPath)
	if err != nil {
		return err
	}
	defer alloc.Close()

	browserCtx, cancelBrowser := chromedp.NewContext(alloc.ctx)
	defer cancelBrowser()

	runCtx, cancelRun := context.WithTimeout(browserCtx, overallTimeout)
	defer cancelRun()

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString(htmlContent)

	var pdfBytes []byte
	err = chromedp.Run(runCtx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(waitForMermaid),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(a4WidthInches).
				WithPaperHeight(a4HeightInches).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBytes = buf
			return nil
		}),
	)
	if err != nil {
		if IsWindowsExecutable(browserPath) && looksLikeNetworkError(err) {
			return fmt.Errorf("%w\n\n%s", err, wslFirewallHelp())
		}
		return fmt.Errorf("generando PDF: %w", err)
	}

	if err := os.WriteFile(outputPath, pdfBytes, 0o644); err != nil {
		return fmt.Errorf("escribiendo PDF en %s: %w", outputPath, err)
	}
	return nil
}

// allocator wraps a chromedp allocator context plus any auxiliary cleanup
// (e.g. terminating a manually launched browser process on WSL).
type allocator struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (a *allocator) Close() { a.cancel() }

func newAllocator(parent context.Context, browserPath string) (*allocator, error) {
	if IsWindowsExecutable(browserPath) {
		return newWSLAllocator(parent, browserPath)
	}
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.Flag("no-sandbox", true),
	)
	ctx, cancel := chromedp.NewExecAllocator(parent, opts...)
	return &allocator{ctx: ctx, cancel: cancel}, nil
}

// looksLikeNetworkError reports whether err appears to come from a failed
// TCP/WS dial. Used to surface a targeted hint for WSL → Windows firewall.
func looksLikeNetworkError(err error) bool {
	msg := err.Error()
	for _, marker := range []string{"i/o timeout", "connection refused", "could not dial", "no route to host"} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// wslFirewallHelp returns user-facing guidance for the most common
// failure mode when chromedp connects to a Windows chrome.exe from WSL:
// Windows Defender Firewall categorising the vEthernet (WSL) interface
// as Public and dropping inbound traffic to chrome.exe.
func wslFirewallHelp() string {
	return strings.Join([]string{
		"El navegador arrancó en Windows pero la comunicación DevTools desde WSL falló.",
		"El Windows Firewall suele bloquear conexiones entrantes a chrome.exe en la red vEthernet (WSL).",
		"",
		"Opciones para resolverlo (ordenadas por simplicidad):",
		"  1. Instalar un navegador Chromium dentro de WSL (recomendado):",
		"       sudo apt install chromium-browser",
		"  2. Habilitar networkingMode=mirrored en %USERPROFILE%\\.wslconfig (Windows 11 22H2+) y reiniciar WSL.",
		"  3. Permitir chrome.exe inbound en Windows Firewall para la red de WSL",
		"     (New-NetFirewallRule en PowerShell con privilegios de administrador).",
	}, "\n")
}

// waitForMermaid polls the rendered document until each .mermaid container
// has an <svg> child, or until mermaidTimeout is exceeded. If the document
// contains no mermaid blocks, it returns immediately.
func waitForMermaid(ctx context.Context) error {
	const script = `(() => {
		const blocks = document.querySelectorAll('.mermaid');
		if (blocks.length === 0) return true;
		return document.querySelectorAll('.mermaid svg').length >= blocks.length;
	})()`

	deadline := time.Now().Add(mermaidTimeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		var done bool
		if err := chromedp.Evaluate(script, &done).Do(ctx); err != nil {
			return fmt.Errorf("evaluando estado de mermaid: %w", err)
		}
		if done {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("timeout esperando a que mermaid termine de renderizar")
}
