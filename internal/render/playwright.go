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

	// Add extra args for better performance and stability
	extraArgs := []string{
		"--disable-web-security",
		"--disable-features=VizDisplayCompositor",
		"--memory-pressure-off",
		"--max_old_space_size=4096",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-ipc-flooding-protection",
		"--disable-background-networking",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-dev-shm-usage",
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

	// Set HTML content with increased timeout and patience for large content
	logs.Infof("Setting HTML content (%d characters)", len(html))
	if err := page.SetContent(html, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(60000), // 60 seconds timeout for heavy content
	}); err != nil {
		logs.Errorf("Failed to set page content: %v", err)
		return nil, fmt.Errorf("failed to set page content: %w", err)
	}

	// More thorough waiting strategy for dynamic content like KaTeX
	logs.Info("Waiting for content to load...")

	// Wait for DOM to be ready
	if _, err := page.WaitForSelector("body", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000), // 10 seconds
	}); err != nil {
		logs.Warnf("Could not wait for body selector: %v", err)
	}

	// Check if page has KaTeX elements that need to be processed
	pageWaitResult, err := page.WaitForSelector("body", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		logs.Warnf("Could not find body element: %v", err)
		pageWaitResult = nil
	}

	if pageWaitResult != nil {
		hasKaTeXPattern, checkErr := page.Evaluate(`() => {
			const bodyText = document.body.innerHTML;
			// Check for KaTeX delimiter patterns in the content
			const hasMathPatterns = /\\\$[^\\\$].*\\\$|\\\$\{2}[^\\\$].*\\\$\{2}|\\\\\\([^)]*\\\\\\)|\\\\\\[[^\\]]*\\\\\\]/.test(bodyText);
			return hasMathPatterns;
		}`)

		if checkErr != nil {
			logs.Warnf("Could not check for KaTeX patterns: %v", checkErr)
			hasKaTeXPattern = true // Assume there might be, to be safe
		}

		if val, ok := hasKaTeXPattern.(bool); ok && val {
			logs.Info("Math formulas detected, waiting for KaTeX rendering...")

			// Attempt manual KaTeX rendering if auto-render didn't happen properly
			// (this addresses the scenario where DOMContentLoaded already fired)
			_, manualRenderErr := page.Evaluate(`() => {
				if (window.renderMath && !window.katexRenderingComplete) {
					window.renderMath();
					// Give a brief moment for the render to process
					window.katexRenderingComplete = true;
				} else if (window.renderMath) {
					// If renderMath exists and rendering isn't complete, call it
					window.renderMath();
				}
			}`)
			if manualRenderErr != nil {
				logs.Warnf("Could not trigger manual KaTeX rendering: %v", manualRenderErr)
			}

			// Wait specifically for KaTeX rendering to complete and elements to appear
			katexWaitResult, katexErr := page.WaitForFunction(`() => {
				// Check if KaTeX rendering has completed
				const katexReady = window.katexRenderingComplete === true;
				const katexElements = document.querySelectorAll('.katex').length;
				return katexReady && katexElements > 0;
			}`, playwright.PageWaitForFunctionOptions{
				Timeout: playwright.Float(30000), // 30 seconds for math rendering
			})
			if katexErr != nil {
				logs.Warnf("Math elements might still be processing after timeout: %v", katexErr)
				// Still proceed, KaTeX might be partially completed
			} else {
				logs.Infof("Successfully waited for KaTeX rendering, result: %v", katexWaitResult)
			}
		} else {
			logs.Info("No math formulas detected in content")
		}
	}

	// Additional wait to ensure all elements render properly
	page.WaitForTimeout(2000)

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
