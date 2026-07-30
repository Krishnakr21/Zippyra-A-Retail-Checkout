package main

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"
)

type Repository interface {
	InsertSalesEvent(ctx context.Context, s *SalesEvent) error
	InsertOrderItemEvents(ctx context.Context, items []*OrderItemEvent) error
	InsertFunnelEvent(ctx context.Context, f *FunnelEvent) error
	IncrementTransactionHourly(ctx context.Context, storeID string, ts time.Time) error

	GetSales(ctx context.Context, storeID, dateFrom, dateTo, granularity string) ([]SalesMetricPeriod, error)
	GetTopProducts(ctx context.Context, storeID, dateFrom, dateTo string, limit int) ([]TopProductItem, error)
	GetFunnel(ctx context.Context, storeID, dateFrom, dateTo string) ([]FunnelStageMetric, error)
	GetPeakHours(ctx context.Context, storeID string, weeksLookback, throughputPerHour int) ([]PeakHourCell, error)
	GetChainSummary(ctx context.Context, chainID, dateFrom, dateTo string) (*ChainSummaryResponse, error)
}

type MemoryRepository struct {
	mu           sync.RWMutex
	sales        map[string]*SalesEvent      // order_id -> latest SalesEvent (ReplacingMergeTree dedup simulation)
	orderItems   map[string]*OrderItemEvent  // order_id+barcode -> latest OrderItemEvent
	funnelEvents []*FunnelEvent
	hourlyCounts map[string]int              // store_id:day_of_week:hour:date_str -> count
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		sales:        make(map[string]*SalesEvent),
		orderItems:   make(map[string]*OrderItemEvent),
		funnelEvents: []*FunnelEvent{},
		hourlyCounts: make(map[string]int),
	}
}

func (m *MemoryRepository) InsertSalesEvent(ctx context.Context, s *SalesEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulates ReplacingMergeTree(event_time): if order_id exists, keep row with latest event_time
	existing, ok := m.sales[s.OrderID]
	if !ok || s.EventTime.After(existing.EventTime) || s.EventTime.Equal(existing.EventTime) {
		cp := *s
		m.sales[s.OrderID] = &cp
	}
	return nil
}

func (m *MemoryRepository) InsertOrderItemEvents(ctx context.Context, items []*OrderItemEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range items {
		key := item.OrderID + ":" + item.Barcode
		existing, ok := m.orderItems[key]
		if !ok || item.EventDate.After(existing.EventDate) || item.EventDate.Equal(existing.EventDate) {
			cp := *item
			m.orderItems[key] = &cp
		}
	}
	return nil
}

func (m *MemoryRepository) InsertFunnelEvent(ctx context.Context, f *FunnelEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := *f
	m.funnelEvents = append(m.funnelEvents, &cp)
	return nil
}

func (m *MemoryRepository) IncrementTransactionHourly(ctx context.Context, storeID string, ts time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dayOfWeek := int(ts.Weekday())
	hour := ts.Hour()
	dateStr := ts.Format("2006-01-02")

	key := storeID + ":" + string(rune(dayOfWeek)) + ":" + string(rune(hour)) + ":" + dateStr
	m.hourlyCounts[key]++
	return nil
}

func (m *MemoryRepository) GetSales(ctx context.Context, storeID, dateFrom, dateTo, granularity string) ([]SalesMetricPeriod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	periodMap := make(map[string]*SalesMetricPeriod)

	for _, s := range m.sales {
		if storeID != "" && s.StoreID != storeID {
			continue
		}
		dateStr := s.EventDate.Format("2006-01-02")
		if (dateFrom != "" && dateStr < dateFrom) || (dateTo != "" && dateStr > dateTo) {
			continue
		}

		periodKey := s.EventDate.Format("2006-01-02")
		if granularity == "month" {
			periodKey = s.EventDate.Format("2006-01")
		}

		p, ok := periodMap[periodKey]
		if !ok {
			p = &SalesMetricPeriod{Period: periodKey}
			periodMap[periodKey] = p
		}
		p.RevenuePaise += s.TotalPaise
		p.DiscountPaise += s.DiscountPaise
		p.OrderCount++
	}

	var result []SalesMetricPeriod
	for _, p := range periodMap {
		result = append(result, *p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Period < result[j].Period
	})

	return result, nil
}

func (m *MemoryRepository) GetTopProducts(ctx context.Context, storeID, dateFrom, dateTo string, limit int) ([]TopProductItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}
	prodMap := make(map[string]*TopProductItem)

	for _, item := range m.orderItems {
		if storeID != "" && item.StoreID != storeID {
			continue
		}

		p, ok := prodMap[item.Barcode]
		if !ok {
			p = &TopProductItem{Barcode: item.Barcode, ProductName: item.ProductName}
			prodMap[item.Barcode] = p
		}
		p.QtySold += int64(item.Qty)
		p.TotalRevenuePaise += item.LineTotalPaise
	}

	var result []TopProductItem
	for _, p := range prodMap {
		result = append(result, *p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalRevenuePaise > result[j].TotalRevenuePaise
	})

	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MemoryRepository) GetFunnel(ctx context.Context, storeID, dateFrom, dateTo string) ([]FunnelStageMetric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect unique session_ids per stage
	stageSessions := make(map[string]map[string]bool)
	for _, stage := range AllFunnelStages {
		stageSessions[stage] = make(map[string]bool)
	}

	for _, f := range m.funnelEvents {
		if storeID != "" && f.StoreID != storeID {
			continue
		}
		if set, ok := stageSessions[f.Stage]; ok && f.SessionID != "" {
			set[f.SessionID] = true
		}
	}

	var result []FunnelStageMetric
	var prevCount int64 = 0

	for idx, stage := range AllFunnelStages {
		count := int64(len(stageSessions[stage]))
		var conversion float64 = 100.0
		if idx > 0 {
			if prevCount > 0 {
				conversion = (float64(count) / float64(prevCount)) * 100.0
			} else {
				conversion = 0.0
			}
		}

		result = append(result, FunnelStageMetric{
			Stage:                         stage,
			SessionCount:                  count,
			ConversionFromPreviousPercent: math.Round(conversion*10) / 10,
		})
		prevCount = count
	}

	return result, nil
}

func (m *MemoryRepository) GetPeakHours(ctx context.Context, storeID string, weeksLookback, throughputPerHour int) ([]PeakHourCell, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if weeksLookback <= 0 {
		weeksLookback = 4
	}
	if throughputPerHour <= 0 {
		throughputPerHour = 20
	}

	// 7 x 24 grid
	grid := make(map[uint8]map[uint8]int64)
	for dow := uint8(0); dow < 7; dow++ {
		grid[dow] = make(map[uint8]int64)
	}

	for key, count := range m.hourlyCounts {
		// Key format: storeID:dow:hour:dateStr
		var sID string
		var dow, hr uint8
		// Aggregate counts by dow and hour
		_ = key
		_ = sID
		_ = dow
		_ = hr
		_ = count
	}

	// Calculate cell metrics across 7 days x 24 hours
	var result []PeakHourCell
	for dow := uint8(0); dow < 7; dow++ {
		for hr := uint8(0); hr < 24; hr++ {
			var totalTx int64 = 0
			for _, s := range m.sales {
				if storeID != "" && s.StoreID != storeID {
					continue
				}
				if uint8(s.EventTime.Weekday()) == dow && uint8(s.EventTime.Hour()) == hr {
					totalTx++
				}
			}

			avgPerWeek := float64(totalTx) / float64(weeksLookback)
			recommended := CalculateRecommendedStaff(avgPerWeek, throughputPerHour)

			result = append(result, PeakHourCell{
				DayOfWeek:              dow,
				Hour:                   hr,
				AvgTransactionsPerWeek: math.Round(avgPerWeek*10) / 10,
				RecommendedStaff:       recommended,
			})
		}
	}

	return result, nil
}

func (m *MemoryRepository) GetChainSummary(ctx context.Context, chainID, dateFrom, dateTo string) (*ChainSummaryResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storeMap := make(map[string]*StoreSalesSummary)
	var totalRevenue int64 = 0
	var totalOrders int64 = 0

	for _, s := range m.sales {
		if chainID != "" && s.ChainID != chainID {
			continue
		}

		st, ok := storeMap[s.StoreID]
		if !ok {
			st = &StoreSalesSummary{
				StoreID:   s.StoreID,
				StoreName: "Store " + s.StoreID[:min(6, len(s.StoreID))],
			}
			storeMap[s.StoreID] = st
		}
		st.RevenuePaise += s.TotalPaise
		st.OrderCount++

		totalRevenue += s.TotalPaise
		totalOrders++
	}

	var byStore []StoreSalesSummary
	for _, st := range storeMap {
		byStore = append(byStore, *st)
	}

	return &ChainSummaryResponse{
		TotalRevenuePaise: totalRevenue,
		TotalOrders:       totalOrders,
		ByStore:           byStore,
	}, nil
}
