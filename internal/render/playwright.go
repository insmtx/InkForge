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

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-web-security",
			"--disable-features=VizDisplayCompositor",
			"--memory-pressure-off",
			"--max_old_space_size=4096",
		},
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

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

	// Set HTML content
	logs.Infof("Setting HTML content (%d characters)", len(html))
	if err := page.SetContent(html, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}); err != nil {
		logs.Errorf("Failed to set page content: %v", err)
		return nil, fmt.Errorf("failed to set page content: %w", err)
	}

	// Wait for the content to fully load
	logs.Info("Waiting for content to load...")
	if _, err := page.WaitForSelector("body", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		logs.Warnf("Could not wait for body selector: %v", err)
	}

	// Additional wait to ensure resources load properly
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
