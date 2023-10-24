package sqlstorage

import (
	"fmt"
	"strings"
	"time"

	"github.com/nianticlabs/modron/src/model"
)

func sqlFilterFromModel(filter model.StorageFilter) string {
	whereList := []string{}
	if len(filter.ResourceGroupNames) > 0 {
		rgns := []string{}
		for _, rgn := range filter.ResourceGroupNames {
			if rgn != "" {
				rgns = append(rgns, fmt.Sprintf("resourcegroupname = '%s'", rgn))
			}
		}
		if len(rgns) > 0 {
			whereList = append(whereList, fmt.Sprintf("(%s)", strings.Join(rgns, " OR ")))
		}
	}
	if len(filter.ResourceTypes) > 0 {
		rts := []string{}
		for _, rt := range filter.ResourceTypes {
			if rt != "" {
				rts = append(rts, fmt.Sprintf("resourcetype = '%s'", rt))
			}
		}
		if len(rts) > 0 {
			whereList = append(whereList, fmt.Sprintf("(%s)", strings.Join(rts, " OR ")))
		}
	}
	if len(filter.ResourceNames) > 0 {
		rns := []string{}
		for _, rt := range filter.ResourceNames {
			if rt != "" {
				rns = append(rns, fmt.Sprintf("resourcename = '%s'", rt))
			}
		}
		if len(rns) > 0 {
			whereList = append(whereList, fmt.Sprintf("(%s)", strings.Join(rns, " OR ")))
		}
	}
	if !filter.StartTime.IsZero() && filter.TimeOffset > 0 {
		whereList = append(whereList, fmt.Sprintf("recordtime >= '%s' AND recordtime <= '%s'", filter.StartTime.Format(time.RFC3339), filter.StartTime.Add(filter.TimeOffset).Format(time.RFC3339)))
	}
	whereClause := ""
	if len(whereList) > 0 {
		whereClause = strings.Join(whereList, " AND ")
	}
	return whereClause
}

func limitClause(filter model.StorageFilter) string {
	if filter.Limit > 0 {
		return fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	return ""
}
