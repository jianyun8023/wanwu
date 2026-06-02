package orm

import (
	"fmt"
	"strconv"

	"github.com/UnicomAI/wanwu/pkg/util"
)

const redisStatsExpireSeconds = 30 * 24 * 3600 // 30天

type StatisticOverviewItem struct {
	Value            float32 `json:"value"`
	PeriodOverPeriod float32 `json:"periodOverPeriod"`
}

func calculatePoP(current, previous float32) float32 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", ((current-previous)/previous)*100), 32)
	return float32(value)
}

func calculateFailureRate(failureCount, totalCount int32) float32 {
	if totalCount == 0 {
		return 0
	}
	return float32(failureCount) / float32(totalCount) * 100
}

func calculateAvg(totalCosts int64, successCount int32) float32 {
	if successCount <= 0 {
		return 0
	}
	return float32(totalCosts) / float32(successCount)
}

type ChartLineValueProvider[T any] func(item T) map[string]float32

func buildChartLines[T any](stats []T, dates []string, getDate func(T) string, getValues ChartLineValueProvider[T], lineNames []string) []StatisticChartLine {
	dateMap := make(map[string]T)
	for _, stat := range stats {
		dateMap[getDate(stat)] = stat
	}

	lines := make([]StatisticChartLine, len(lineNames))
	for i, name := range lineNames {
		lines[i] = StatisticChartLine{
			Name:  name,
			Items: make([]StatisticChartLineItem, 0, len(dates)),
		}
	}

	for _, date := range dates {
		if stat, ok := dateMap[date]; ok {
			values := getValues(stat)
			for i := range lines {
				lines[i].Items = append(lines[i].Items, StatisticChartLineItem{
					Key:   date,
					Value: values[lines[i].Name],
				})
			}
		} else {
			for i := range lines {
				lines[i].Items = append(lines[i].Items, StatisticChartLineItem{
					Key:   date,
					Value: 0,
				})
			}
		}
	}

	return lines
}

func buildDateRange(startDate, endDate string) ([]string, error) {
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return nil, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return nil, err
	}
	return util.DateRange(startTs, endTs), nil
}
