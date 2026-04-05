package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Wide fan-in graph: input ─► fetch-weather ─┐
//                          ─► fetch-traffic  ─┼─► aggregate-report
//                          ─► fetch-events   ─┘
// All three source nodes depend only on input, so they execute in parallel.

func TestGetDashboard(t *testing.T) {
	type args struct {
		city string
	}

	testCases := []struct {
		name  string
		args  args
		check func(t *testing.T, resp getDashboardResponse)
	}{
		{
			name: "returns aggregated report with all three data sources",
			args: args{city: "Moscow"},
			check: func(t *testing.T, resp getDashboardResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "Moscow", resp.Report.Weather.City)
				require.Equal(t, "Sunny", resp.Report.Weather.Condition)
				require.Equal(t, "Moscow", resp.Report.Traffic.City)
				require.Equal(t, 42, resp.Report.Traffic.CongestionPct)
				require.Equal(t, "Moscow", resp.Report.Events.City)
				require.NotEmpty(t, resp.Report.Events.Events)
			},
		},
		{
			name: "uses default city when city param is omitted",
			args: args{city: ""},
			check: func(t *testing.T, resp getDashboardResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "Moscow", resp.Report.Weather.City)
				require.Equal(t, "Moscow", resp.Report.Traffic.City)
			},
		},
		{
			name: "returns 500 when events source fails",
			args: args{city: "error"},
			check: func(t *testing.T, resp getDashboardResponse) {
				require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := getDashboard(t, tc.args.city)
			tc.check(t, resp)
		})
	}
}

type dashboardWeather struct {
	City        string `json:"city"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
}

type dashboardTraffic struct {
	City          string `json:"city"`
	CongestionPct int    `json:"congestion_pct"`
	Incidents     int    `json:"incidents"`
}

type dashboardEvents struct {
	City   string   `json:"city"`
	Events []string `json:"events"`
}

type dashboardReport struct {
	Weather dashboardWeather `json:"weather"`
	Traffic dashboardTraffic `json:"traffic"`
	Events  dashboardEvents  `json:"events"`
}

type getDashboardResponse struct {
	StatusCode int
	Report     dashboardReport
}

func getDashboard(t *testing.T, city string) getDashboardResponse {
	t.Helper()

	url := shortcutURL + "/run/dashboard/dashboard/get-dashboard"
	if city != "" {
		url += "?city=" + city
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := getDashboardResponse{StatusCode: resp.StatusCode}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Report))
	}

	return result
}
