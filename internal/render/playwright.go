package render

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/ygpkg/yg-go/logs"
)

// PagePool manages a pool of reusable browser pages for efficient rendering
type PagePool struct {
	pool      chan playwright.Page
	browser   playwright.Browser
	context   playwright.BrowserContext
	maxPages  int
	mu        sync.Mutex
	usedCount int
}

// PlaywrightRenderer handles browser-based rendering using Playwright
type PlaywrightRenderer struct {
	pw       *playwright.Playwright
	browser  playwright.Browser
	context  playwright.BrowserContext
	pagePool *PagePool
}

// NewPlaywrightRenderer initializes the Playwright renderer with a page pool
func NewPlaywrightRenderer() (*PlaywrightRenderer, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run Playwright: %w", err)
	}

	// Add extra args for better performance, stability and network resilience
	extraArgs := []string{
		"--disable-web-security",
		"--disable-features=VizDisplayCompositor",
		"--memory-pressure-off",
		"--max_old_space_size=4096",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-ipc-flooding-protection",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-dev-shm-usage",
		"--disable-extensions",
		"--disable-plugins",
		"--no-default-browser-check",
		"--disable-features=TranslateUI,BlinkGenPropertyTrees",
		"--disable-hang-monitor",
		"--disable-prompt-on-repost",
		"--disable-sync",
		"--no-first-run",
		"--disable-blink-features=AutomationControlled",
		"--disable-webgl",
		"--disable-logging",
		"--ignore-certificate-errors",
		"--ignore-certificate-errors-skip-list",
		"--disable-features=VizDisplayCompositor",
		"--no-sandbox",            // Important for containerized environments
		"--disable-dev-shm-usage", // Overcome limited resource problems
		"--disable-gpu",           // Disable GPU hardware acceleration
		"--remote-debugging-port=9222",
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args:     extraArgs,
		Timeout:  playwright.Float(60000), // Increase startup timeout
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Increase context creation timeout for complex rendering
	context, err := browser.NewContext()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	// Create page pool with 10 pages initially
	pagePool, err := newPagePool(context, browser, 10)
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to initialize page pool: %w", err)
	}

	renderer := &PlaywrightRenderer{
		pw:       pw,
		browser:  browser,
		context:  context,
		pagePool: pagePool,
	}

	logs.Infof("Successfully initialized Playwright renderer with page pool of %d pages", pagePool.maxPages)
	return renderer, nil
}

// newPagePool creates a new pool with the specified number of pre-loaded pages
func newPagePool(context playwright.BrowserContext, browser playwright.Browser, maxPages int) (*PagePool, error) {
	pool := make(chan playwright.Page, maxPages)
	pagePool := &PagePool{
		pool:     pool,
		browser:  browser,
		context:  context,
		maxPages: maxPages,
	}

	// Pre-create pages for the pool
	logs.Infof("Pre-loading %d browser pages for pool...", maxPages)
	for i := 0; i < maxPages; i++ {
		page, err := context.NewPage()
		if err != nil {
			return nil, fmt.Errorf("failed to create page %d: %w", i, err)
		}

		// Set a blank initial state for the page
		if _, err := page.Goto("about:blank"); err != nil {
			logs.Warnf("Could not navigate new page to blank: %v", err)
		}

		// Ensure viewport is set to standard size
		if err := page.SetViewportSize(1200, 800); err != nil {
			logs.Warnf("Could not set initial viewport: %v", err)
		}

		select {
		case pagePool.pool <- page:
			logs.Debugf("Added pre-loaded page to pool")
		default:
			// Pool channel is full - shouldn't happen with buffered channel
			logs.Warnf("Pool channel is full, attempting to close page")
			page.Close()
		}
	}

	logs.Infof("Successfully created and populated page pool with %d pages", maxPages)
	return pagePool, nil
}

// acquirePage gets a page from the pool, with timeout
func (pp *PagePool) acquirePage(timeout time.Duration) (playwright.Page, error) {
	pp.mu.Lock()
	pp.usedCount++
	currentUsed := pp.usedCount
	pp.mu.Unlock()

	logs.Debugf("Acquiring page from pool, currently in use count: %d", currentUsed)

	select {
	case page := <-pp.pool:
		return page, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout acquiring page from pool after %v", timeout)
	}
}

// releasePage returns a page to the pool, but resets its state first
func (pp *PagePool) releasePage(page playwright.Page) error {
	// Reset page state by navigating to blank and setting default viewport
	if _, err := page.Goto("about:blank"); err != nil {
		logs.Warnf("Failed to reset page to blank state: %v", err)
		// If navigation fails, recreate the page
		newPage, err := pp.context.NewPage()
		if err != nil {
			return fmt.Errorf("failed to create replacement page: %w", err)
		}
		page.Close()
		page = newPage
	} else {
		// Reset viewport
		if err := page.SetViewportSize(1200, 800); err != nil {
			logs.Warnf("Could not reset viewport on page: %v", err)
		}
	}

	pp.mu.Lock()
	pp.usedCount--
	currentUsed := pp.usedCount
	pp.mu.Unlock()

	logs.Debugf("Releasing page to pool, currently in use count: %d", currentUsed)

	// Try to return page to pool
	select {
	case pp.pool <- page:
		return nil
	default:
		// Pool is full - close the page
		logs.Debugf("Page pool is full, closing page")
		page.Close()
		return nil
	}
}

// RenderImage renders the given HTML content to an image using pages from the pool
func (r *PlaywrightRenderer) RenderImage(ctx context.Context, html string, width, height int, scale float64, imgFormat string) ([]byte, error) {
	logs.Infof("Starting image rendering process - Width: %d, Height: %d, Scale: %.2f, Format: %s", width, height, scale, imgFormat)

	// Acquire a page from pool with timeout
	page, err := r.pagePool.acquirePage(30 * time.Second)
	if err != nil {
		logs.Errorf("Failed to acquire page from pool: %v", err)
		return nil, fmt.Errorf("failed to acquire page: %w", err)
	}

	defer func() {
		if closeErr := r.pagePool.releasePage(page); closeErr != nil {
			logs.Errorf("Warning: Failed to return page to pool: %v", closeErr)
		} else {
			logs.Debugf("Successfully returned page to pool after rendering")
		}
	}()

	// Set viewport size
	logs.Infof("Setting viewport size to %dx%d", width, height)
	if err := page.SetViewportSize(width, height); err != nil {
		logs.Errorf("Failed to set viewport size: %v", err)
		return nil, fmt.Errorf("failed to set viewport: %w", err)
	}

	// Set HTML content but with more tolerant options to prevent timeout during external resource loading
	logs.Infof("Setting HTML content (%d characters)", len(html))

	// Navigate to a blank page first to ensure fresh state
	if _, err := page.Goto("about:blank"); err != nil {
		logs.Warnf("Could not navigate to blank page: %v", err)
	}

	// Set content with more relaxed waitUntil condition to avoid hanging on external resource loads
	if err := page.SetContent(html, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded, // Changed from networkidle to domcontentload to avoid timeout on network resources
		Timeout:   playwright.Float(60000),                   // 60 seconds but with faster resolution
	}); err != nil {
		logs.Errorf("Failed to set page content: %v", err)
		return nil, fmt.Errorf("failed to set page content: %w", err)
	}

	// More thorough waiting strategy for dynamic content like KaTeX
	logs.Info("Waiting for content to load...")

	// Wait for DOM to be ready
	if _, err := page.WaitForSelector("body", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000), // 10 seconds for DOM to be ready
	}); err != nil {
		logs.Warnf("Could not wait for body selector: %v", err)
	}

	// Check if page has KaTeX elements that need to be processed
	hasKaTeXPatternResult, patternCheckErr := page.Evaluate(`() => {
		const bodyInnerHTML = document.body.innerHTML;
		// Check for KaTeX delimiter patterns in the content
		const hasDollarFormula = bodyInnerHTML.includes('$');
		const hasEscapeParens = bodyInnerHTML.includes('\\(') || bodyInnerHTML.includes('\\)');
		const hasEscapeBrackets = bodyInnerHTML.includes('\\[') || bodyInnerHTML.includes('\\]');
		
		return hasDollarFormula || hasEscapeParens || hasEscapeBrackets;
	}`)

	if patternCheckErr != nil {
		logs.Warnf("Could not check for KaTeX patterns: %v", patternCheckErr)
		hasKaTeXPatternResult = true // Default to true if we can't check
	}

	hasKaTeX := false
	if val, ok := hasKaTeXPatternResult.(bool); ok {
		hasKaTeX = val
	}

	if hasKaTeX {
		logs.Info("Math formulas detected, waiting for KaTeX rendering...")

		// Allow some time for CDN resources to load (but not indefinitely)
		page.WaitForTimeout(5000) // Give it maximum 5 seconds to load KaTeX resources

		// Attempt manual KaTeX rendering if available
		_, manualRenderErr := page.Evaluate(`() => {
			// Make sure renderMath is available or retry
			if (typeof renderMathInElement !== 'undefined' && document.body) {
				renderMathInElement(document.body, {
					delimiters: [
						{left: "$$", right: "$$", display: true},
						{left: "$", right: "$", display: false},
						{left: "\\(", right: "\\)", display: false},
						{left: "\\[", right: "\\]", display: true}
					],
					ignoredTags: ['script', 'noscript', 'style', 'textarea', 'pre', 'code']
				});
				window.katexRenderingComplete = true;
			} else if (window.katex) {
				window.katexRenderingComplete = true; // Assume rendering might be in progress
			}
		}`)
		if manualRenderErr != nil {
			logs.Warnf("Could not trigger manual KaTeX rendering: %v", manualRenderErr)
		}

		// Wait up to 10 more seconds specifically for KaTeX rendered elements
		katexWaitStart := time.Now()
		katexFinished := false

		// Use a polling approach instead of waiting forever
		for i := 0; i < 10; i++ { // 10 iterations * 1 sec = 10 seconds max
			result, evalErr := page.Evaluate(`() => {
				// Check for rendered KaTeX elements or completion flag
				const katexElements = document.querySelectorAll('.katex').length;
				const completeFlag = window.katexRenderingComplete === true;
				return { count: katexElements, ready: completeFlag };
			}`)

			if evalErr == nil && result != nil {
				if resultMap, ok := result.(map[string]interface{}); ok {
					count := 0
					if cnt, ok := resultMap["count"].(float64); ok {
						count = int(cnt)
					}

					if ready, ok := resultMap["ready"].(bool); ok && ready || count > 0 {
						katexFinished = true
						logs.Infof("KaTeX rendered %d element(s)", count)
						break
					}
				}
			}

			time.Sleep(1 * time.Second) // Brief pause between checks
		}

		if !katexFinished {
			logs.Warnf("Math elements might still be processing after max timeout: %s", time.Since(katexWaitStart))
		}
	} else {
		logs.Info("No math formulas detected in content")
	}

	// Final wait to ensure visuals are stable
	page.WaitForTimeout(1000)

	// Determine screenshot options based on image format
	var screenshotOptions playwright.PageScreenshotOptions

	switch imgFormat {
	case "jpeg", "jpg":
		screenshotOptions = playwright.PageScreenshotOptions{
			Type:     playwright.ScreenshotTypeJpeg,
			Quality:  playwright.Int(90),
			FullPage: playwright.Bool(false),
		}
	case "png":
		screenshotOptions = playwright.PageScreenshotOptions{
			Type:     playwright.ScreenshotTypePng,
			FullPage: playwright.Bool(false),
		}
	default:
		screenshotOptions = playwright.PageScreenshotOptions{
			Type:     playwright.ScreenshotTypePng,
			FullPage: playwright.Bool(false),
		}
	}

	logs.Infof("Taking screenshot with options - Type: %s", screenshotOptions.Type)

	// Take the screenshot
	screenshot, err := page.Screenshot(screenshotOptions)
	if err != nil {
		logs.Errorf("Failed to take screenshot: %v", err)
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	logs.Infof("Successfully captured screenshot, size: %d bytes", len(screenshot))
	return screenshot, nil
}

// Close releases the resources held by the renderer
func (r *PlaywrightRenderer) Close() error {
	var errs []error

	// Close all pages in the pool
	if r.pagePool != nil {
		// Drain and close all pages in the pool
		if r.pagePool.pool != nil {
			// Get all pages from the pool and close them
			drainPages := make([]playwright.Page, 0, len(r.pagePool.pool))
			for len(r.pagePool.pool) > 0 {
				select {
				case page := <-r.pagePool.pool:
					drainPages = append(drainPages, page)
				default:
					// If nothing to drain anymore, break
					break
				}
			}

			// Close all drained pages
			for _, page := range drainPages {
				if err := page.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close browser page: %w", err))
				}
			}
		}
	}

	if err := r.context.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close browser context: %w", err))
	}

	if err := r.browser.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close browser: %w", err))
	}

	if err := r.pw.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("failed to stop Playwright: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during Playwright cleanup: %v", errs)
	}

	logs.Infof("Successfully closed Playwright renderer")
	return nil
}

// RenderHTMLToImage is a convenience method that takes an HTML document and returns an image
func (r *PlaywrightRenderer) RenderHTMLToImage(ctx context.Context, html string, width, height int, scale float64, imgFormat string) ([]byte, error) {
	return r.RenderImage(ctx, html, width, height, scale, imgFormat)
}
