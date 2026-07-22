package shared

import (
	"strings"

	"github.com/dujiao-next/internal/modules/reporting"
	"github.com/gin-gonic/gin"
)

// ParseReportingQuery parses the common range contract used by operational reports.
func ParseReportingQuery(c *gin.Context) (reporting.Query, error) {
	from, err := ParseTimeNullable(strings.TrimSpace(c.Query("from")))
	if err != nil {
		return reporting.Query{}, err
	}
	to, err := ParseTimeNullable(strings.TrimSpace(c.Query("to")))
	if err != nil {
		return reporting.Query{}, err
	}
	forceRefresh, err := ParseQueryBool(c, "force_refresh")
	if err != nil {
		return reporting.Query{}, err
	}
	return reporting.Query{
		Range:        strings.TrimSpace(c.DefaultQuery("range", "7d")),
		From:         from,
		To:           to,
		Timezone:     strings.TrimSpace(c.Query("tz")),
		ForceRefresh: forceRefresh,
	}, nil
}
