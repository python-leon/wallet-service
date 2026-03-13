package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 配置参数
type Config struct {
	BaseURL     string
	Concurrency int
	Requests    int
	Duration    time.Duration
	Timeout     time.Duration
}

// 测试结果
type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

// 统计报告
type Report struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	AvgDuration     time.Duration
	P50Duration     time.Duration
	P90Duration     time.Duration
	P95Duration     time.Duration
	P99Duration     time.Duration
	QPS             float64
	StatusCodes     map[int]int64
	Errors          map[string]int64
}

// 转账正确性验证报告
type TransferCorrectnessReport struct {
	InitialTotalBalance   int64
	FinalTotalBalance     int64
	BalanceConserved      bool
	TransferCount         int64
	SuccessTransfers      int64
	FailedTransfers       int64
	WalletBalances        map[string]int64
	NoNegativeBalance     bool
	RaceConditionDetected bool
}

// 钱包响应
type WalletResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID      string `json:"id"`
		Balance int64  `json:"balance"`
	} `json:"data"`
}

// 转账响应
type TransferResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	FromWalletID string `json:"from_wallet_id"`
	ToWalletID   string `json:"to_wallet_id"`
	FromBalance  int64  `json:"from_balance"`
	ToBalance    int64  `json:"to_balance"`
}

// 负载测试器
type LoadTester struct {
	config  Config
	client  *http.Client
	results []Result
	mu      sync.Mutex
}

func NewLoadTester(config Config) *LoadTester {
	return &LoadTester{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		results: make([]Result, 0),
	}
}

// 创建钱包并返回ID
func (lt *LoadTester) createWallet() (string, error) {
	resp, err := lt.client.Post(lt.config.BaseURL+"/wallets", "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var walletResp WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&walletResp); err != nil {
		return "", err
	}
	if !walletResp.Success {
		return "", fmt.Errorf("create wallet failed")
	}
	return walletResp.Data.ID, nil
}

// 获取钱包余额
func (lt *LoadTester) getWalletBalance(walletID string) (int64, error) {
	resp, err := lt.client.Get(lt.config.BaseURL + "/wallets/" + walletID)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var walletResp WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&walletResp); err != nil {
		return 0, err
	}
	if !walletResp.Success {
		return 0, fmt.Errorf("get wallet failed")
	}
	return walletResp.Data.Balance, nil
}

// 存款
func (lt *LoadTester) deposit(walletID string, amount int64) error {
	body := map[string]interface{}{
		"wallet_id": walletID,
		"amount":    amount,
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := lt.client.Post(
		lt.config.BaseURL+"/wallets/deposit",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var walletResp WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&walletResp); err != nil {
		return err
	}
	if !walletResp.Success {
		return fmt.Errorf("deposit failed")
	}
	return nil
}

// 转账测试
func (lt *LoadTester) testTransfer(fromID, toID string, amount int64) (TransferResponse, error) {
	body := map[string]interface{}{
		"from_wallet_id": fromID,
		"to_wallet_id":   toID,
		"amount":         amount,
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := lt.client.Post(
		lt.config.BaseURL+"/wallets/transfer",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return TransferResponse{}, err
	}
	defer resp.Body.Close()

	var transferResp TransferResponse
	if err := json.NewDecoder(resp.Body).Decode(&transferResp); err != nil {
		return TransferResponse{}, err
	}
	return transferResp, nil
}

// 创建钱包测试
func (lt *LoadTester) testCreateWallet() Result {
	start := time.Now()
	resp, err := lt.client.Post(lt.config.BaseURL+"/wallets", "application/json", nil)
	if err != nil {
		return Result{Duration: time.Since(start), Error: err}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return Result{
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
	}
}

// 获取钱包测试
func (lt *LoadTester) testGetWallet(walletID string) Result {
	start := time.Now()
	resp, err := lt.client.Get(lt.config.BaseURL + "/wallets/" + walletID)
	if err != nil {
		return Result{Duration: time.Since(start), Error: err}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return Result{
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
	}
}

// 记录结果
func (lt *LoadTester) recordResult(result Result) {
	lt.mu.Lock()
	lt.results = append(lt.results, result)
	lt.mu.Unlock()
}

// 运行创建钱包负载测试
func (lt *LoadTester) runCreateWalletTest() Report {
	fmt.Println("\n🚀 开始创建钱包负载测试...")
	fmt.Printf("   并发数: %d, 总请求数: %d\n", lt.config.Concurrency, lt.config.Requests)

	var wg sync.WaitGroup
	var counter int64

	start := time.Now()

	for i := 0; i < lt.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				reqNum := atomic.AddInt64(&counter, 1)
				if reqNum > int64(lt.config.Requests) {
					atomic.AddInt64(&counter, -1)
					return
				}
				result := lt.testCreateWallet()
				lt.recordResult(result)
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(start)

	return lt.generateReport(totalDuration)
}

// 运行并发转账正确性测试
func (lt *LoadTester) runTransferCorrectnessTest() (Report, TransferCorrectnessReport) {
	fmt.Println("\n🚀 开始并发转账正确性测试...")
	fmt.Printf("   并发数: %d\n", lt.config.Concurrency)

	correctnessReport := TransferCorrectnessReport{
		WalletBalances: make(map[string]int64),
	}

	// 创建测试钱包
	fmt.Println("   准备测试钱包...")
	numWallets := lt.config.Concurrency * 2
	walletIDs := make([]string, numWallets)
	initialBalances := make(map[string]int64)

	for i := 0; i < numWallets; i++ {
		id, err := lt.createWallet()
		if err != nil {
			fmt.Printf("   创建钱包失败: %v\n", err)
			continue
		}
		walletIDs[i] = id
	}
	fmt.Printf("   已创建 %d 个测试钱包\n", len(walletIDs))

	// 为每个钱包存入初始余额
	initialBalance := int64(1000) // 每个钱包初始余额 1000
	fmt.Printf("   为每个钱包存入初始余额: %d\n", initialBalance)

	for i, id := range walletIDs {
		if id == "" {
			continue
		}
		if err := lt.deposit(id, initialBalance); err != nil {
			fmt.Printf("   存款失败 (钱包 %d): %v\n", i, err)
			continue
		}
		initialBalances[id] = initialBalance
	}

	// 计算初始总余额
	var initialTotal int64
	for _, balance := range initialBalances {
		initialTotal += balance
	}
	correctnessReport.InitialTotalBalance = initialTotal
	fmt.Printf("   初始总余额: %d\n", initialTotal)

	// 并发转账测试
	fmt.Println("   执行并发转账...")

	var wg sync.WaitGroup
	var transferCount int64
	var successCount int64
	var failCount int64
	transferAmount := int64(10) // 每次转账 10

	// 使用通道控制并发
	start := time.Now()

	for i := 0; i < lt.config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < lt.config.Requests/lt.config.Concurrency; j++ {
				// 选择两个不同的钱包进行转账
				fromIdx := (workerID + j*2) % numWallets
				toIdx := (workerID + j*2 + 1) % numWallets

				if walletIDs[fromIdx] == "" || walletIDs[toIdx] == "" {
					continue
				}

				atomic.AddInt64(&transferCount, 1)
				result := lt.testCreateWallet() // 记录一次请求
				lt.recordResult(result)

				transferResp, err := lt.testTransfer(walletIDs[fromIdx], walletIDs[toIdx], transferAmount)
				if err != nil || !transferResp.Success {
					atomic.AddInt64(&failCount, 1)
					continue
				}
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	correctnessReport.TransferCount = transferCount
	correctnessReport.SuccessTransfers = successCount
	correctnessReport.FailedTransfers = failCount

	// 验证最终余额
	fmt.Println("   验证最终余额...")
	var finalTotal int64
	hasNegative := false

	for _, id := range walletIDs {
		if id == "" {
			continue
		}
		balance, err := lt.getWalletBalance(id)
		if err != nil {
			fmt.Printf("   获取钱包 %s 余额失败: %v\n", id, err)
			continue
		}
		correctnessReport.WalletBalances[id] = balance
		finalTotal += balance
		if balance < 0 {
			hasNegative = true
		}
	}

	correctnessReport.FinalTotalBalance = finalTotal
	correctnessReport.BalanceConserved = (initialTotal == finalTotal)
	correctnessReport.NoNegativeBalance = !hasNegative

	// 检测竞态条件
	// 如果余额守恒且没有负余额，则没有竞态条件
	correctnessReport.RaceConditionDetected = !correctnessReport.BalanceConserved || !correctnessReport.NoNegativeBalance

	fmt.Printf("   最终总余额: %d\n", finalTotal)
	fmt.Printf("   余额守恒: %v\n", correctnessReport.BalanceConserved)
	fmt.Printf("   无负余额: %v\n", correctnessReport.NoNegativeBalance)

	return lt.generateReport(totalDuration), correctnessReport
}

// 运行混合负载测试
func (lt *LoadTester) runMixedTest() Report {
	fmt.Println("\n🚀 开始混合负载测试...")
	fmt.Printf("   并发数: %d, 持续时间: %v\n", lt.config.Concurrency, lt.config.Duration)

	// 先创建一些钱包用于测试
	fmt.Println("   准备测试数据...")
	walletIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		resp, err := lt.client.Post(lt.config.BaseURL+"/wallets", "application/json", nil)
		if err != nil {
			fmt.Printf("   准备数据失败: %v\n", err)
			continue
		}
		var walletResp WalletResponse
		if err := json.NewDecoder(resp.Body).Decode(&walletResp); err == nil && walletResp.Success {
			walletIDs[i] = walletResp.Data.ID
		}
		resp.Body.Close()
	}
	fmt.Printf("   已创建 %d 个测试钱包\n", len(walletIDs))

	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	var counter int64

	start := time.Now()

	for i := 0; i < lt.config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					reqNum := atomic.AddInt64(&counter, 1)
					var result Result

					// 根据请求序号选择不同的操作
					// 70% 创建钱包, 20% 获取钱包, 10% 转账
					remainder := reqNum % 10
					switch {
					case remainder < 7:
						result = lt.testCreateWallet()
					case remainder < 9:
						// 随机选择一个钱包ID
						idx := int(reqNum % int64(len(walletIDs)))
						if walletIDs[idx] != "" {
							result = lt.testGetWallet(walletIDs[idx])
						} else {
							result = lt.testCreateWallet()
						}
					default:
						// 转账测试
						fromIdx := int(reqNum % int64(len(walletIDs)))
						toIdx := int((reqNum + 1) % int64(len(walletIDs)))
						if walletIDs[fromIdx] != "" && walletIDs[toIdx] != "" {
							_, err := lt.testTransfer(walletIDs[fromIdx], walletIDs[toIdx], 1)
							if err != nil {
								result = Result{StatusCode: 400, Duration: 0, Error: err}
							} else {
								result = Result{StatusCode: 200, Duration: 0}
							}
						} else {
							result = lt.testCreateWallet()
						}
					}

					lt.recordResult(result)
				}
			}
		}(i)
	}

	// 运行指定时间
	time.Sleep(lt.config.Duration)
	close(stopCh)
	wg.Wait()
	totalDuration := time.Since(start)

	return lt.generateReport(totalDuration)
}

// 生成报告
func (lt *LoadTester) generateReport(totalDuration time.Duration) Report {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	report := Report{
		TotalDuration: totalDuration,
		StatusCodes:   make(map[int]int64),
		Errors:        make(map[string]int64),
	}

	if len(lt.results) == 0 {
		return report
	}

	// 收集延迟数据
	durations := make([]time.Duration, len(lt.results))
	for i, r := range lt.results {
		report.TotalRequests++
		durations[i] = r.Duration

		if r.Error != nil {
			report.FailedRequests++
			errMsg := r.Error.Error()
			report.Errors[errMsg]++
		} else {
			report.SuccessRequests++
			report.StatusCodes[r.StatusCode]++
		}
	}

	// 排序延迟数据
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	// 计算统计数据
	report.MinDuration = durations[0]
	report.MaxDuration = durations[len(durations)-1]

	// 计算平均延迟
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	report.AvgDuration = total / time.Duration(len(durations))

	// 计算百分位延迟
	report.P50Duration = durations[len(durations)*50/100]
	report.P90Duration = durations[len(durations)*90/100]
	report.P95Duration = durations[len(durations)*95/100]
	report.P99Duration = durations[len(durations)*99/100]

	// 计算 QPS
	report.QPS = float64(report.TotalRequests) / totalDuration.Seconds()

	return report
}

// 打印报告
func printReport(report Report) {
	fmt.Println("\n" + "========================================")
	fmt.Println("📊 负载测试报告")
	fmt.Println("========================================")
	fmt.Printf("总请求数:      %d\n", report.TotalRequests)
	fmt.Printf("成功请求:      %d (%.2f%%)\n", report.SuccessRequests, float64(report.SuccessRequests)/float64(report.TotalRequests)*100)
	fmt.Printf("失败请求:      %d (%.2f%%)\n", report.FailedRequests, float64(report.FailedRequests)/float64(report.TotalRequests)*100)
	fmt.Printf("总耗时:        %v\n", report.TotalDuration)
	fmt.Printf("QPS:           %.2f 请求/秒\n", report.QPS)
	fmt.Println()
	fmt.Println("延迟统计:")
	fmt.Printf("  最小:        %v\n", report.MinDuration)
	fmt.Printf("  最大:        %v\n", report.MaxDuration)
	fmt.Printf("  平均:        %v\n", report.AvgDuration)
	fmt.Printf("  P50:         %v\n", report.P50Duration)
	fmt.Printf("  P90:         %v\n", report.P90Duration)
	fmt.Printf("  P95:         %v\n", report.P95Duration)
	fmt.Printf("  P99:         %v\n", report.P99Duration)

	if len(report.StatusCodes) > 0 {
		fmt.Println("\n状态码分布:")
		for code, count := range report.StatusCodes {
			fmt.Printf("  %d: %d (%.2f%%)\n", code, count, float64(count)/float64(report.TotalRequests)*100)
		}
	}

	if len(report.Errors) > 0 {
		fmt.Println("\n错误分布:")
		for err, count := range report.Errors {
			fmt.Printf("  %s: %d\n", err, count)
		}
	}
	fmt.Println("========================================")
}

// 打印转账正确性报告
func printTransferCorrectnessReport(report TransferCorrectnessReport) {
	fmt.Println("\n" + "========================================")
	fmt.Println("🔍 并发转账正确性验证报告")
	fmt.Println("========================================")
	fmt.Printf("转账次数:          %d\n", report.TransferCount)
	fmt.Printf("成功转账:          %d\n", report.SuccessTransfers)
	fmt.Printf("失败转账:          %d (余额不足等)\n", report.FailedTransfers)
	fmt.Println()
	fmt.Println("余额验证:")
	fmt.Printf("  初始总余额:      %d\n", report.InitialTotalBalance)
	fmt.Printf("  最终总余额:      %d\n", report.FinalTotalBalance)
	fmt.Printf("  余额守恒:        %v\n", report.BalanceConserved)
	fmt.Printf("  无负余额:        %v\n", report.NoNegativeBalance)
	fmt.Println()
	fmt.Println("并发安全性:")
	if report.RaceConditionDetected {
		fmt.Println("  ⚠️  检测到竞态条件!")
	} else {
		fmt.Println("  ✅ 未检测到竞态条件")
	}
	fmt.Println("========================================")
}

func main() {
	// 解析命令行参数
	baseURL := flag.String("url", "http://localhost:8080", "服务基础URL")
	concurrency := flag.Int("c", 10, "并发数")
	requests := flag.Int("n", 1000, "总请求数 (用于创建钱包测试)")
	duration := flag.Duration("d", 30*time.Second, "持续时间 (用于混合测试)")
	timeout := flag.Duration("timeout", 10*time.Second, "请求超时时间")
	testType := flag.String("test", "create", "测试类型: create (创建钱包), mixed (混合测试), transfer (转账正确性)")
	flag.Parse()

	config := Config{
		BaseURL:     *baseURL,
		Concurrency: *concurrency,
		Requests:    *requests,
		Duration:    *duration,
		Timeout:     *timeout,
	}

	fmt.Println("========================================")
	fmt.Println("💪 Wallet Service 负载测试工具")
	fmt.Println("========================================")
	fmt.Printf("目标服务: %s\n", config.BaseURL)
	fmt.Printf("并发数:   %d\n", config.Concurrency)
	fmt.Printf("超时时间: %v\n", config.Timeout)

	// 检查服务是否可用
	fmt.Println("\n检查服务状态...")
	resp, err := http.Get(config.BaseURL + "/health")
	if err != nil {
		fmt.Printf("❌ 无法连接到服务: %v\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	fmt.Println("✅ 服务正常运行")

	// 创建负载测试器
	tester := NewLoadTester(config)

	// 运行测试
	var report Report
	var transferReport TransferCorrectnessReport

	switch *testType {
	case "create":
		report = tester.runCreateWalletTest()
	case "mixed":
		report = tester.runMixedTest()
	case "transfer":
		report, transferReport = tester.runTransferCorrectnessTest()
		printTransferCorrectnessReport(transferReport)
	default:
		fmt.Printf("❌ 未知的测试类型: %s\n", *testType)
		os.Exit(1)
	}

	// 打印报告
	printReport(report)
}
