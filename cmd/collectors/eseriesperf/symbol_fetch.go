package eseriesperf

import (
	"errors"
	"fmt"
	"log/slog"

	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"

	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

// isLegacyFlashCache reports whether the array runs SANtricity < 12.00.
// On those arrays ssd_cache metrics are only available via the SYMbol POST API.
func isLegacyFlashCache(ep *EseriesPerf) bool {
	v, err := goversion.NewVersion(ep.Remote.Version)
	if err != nil {
		return false
	}
	v12, _ := goversion.NewVersion("12.00.0")
	return v.LessThan(v12)
}

// symbolSsdCacheFetch collects ssd_cache statistics on SANtricity < 12.00 arrays
// via the SYMbol POST getFlashCacheStatistics endpoint.
func symbolSsdCacheFetch(ep *EseriesPerf, systemID string, headers map[string]string) ([]gjson.Result, error) {
	endpoint := fmt.Sprintf("%s/storage-systems/%s/flash-cache", ep.Client.APIPath, systemID)
	fc, err := ep.Client.Fetch(endpoint, nil)
	if err != nil {
		// 404 means no SSD cache is configured on this array — treat as no instances.
		if re, ok := errors.AsType[*errs.RestError](err); ok && re.StatusCode == 404 {
			return nil, errs.New(errs.ErrNoInstance, "no SSD cache configured")
		}
		return nil, fmt.Errorf("failed to fetch flash-cache: %w", err)
	}
	if len(fc) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no flash cache found")
	}

	cacheRef := fc[0].Get("flashCacheRef").ClonedString()
	cacheName := fc[0].Get("name").ClonedString()

	if cacheRef == "" {
		return nil, errs.New(errs.ErrNoInstance, "flash cache missing flashCacheRef")
	}

	// Set global labels so the cache name and ref are available on all instances.
	ep.Matrix[ep.Object].SetGlobalLabel("ssd_cache", cacheName)
	ep.Matrix[ep.Object].SetGlobalLabel("ssd_cache_id", cacheRef)

	// Query each controller separately — SYMbol ?controller=auto returns only one
	// controller's data. Calling with a and b gives per-controller instances that
	// match the REST path structure (2 instances per array).
	var results []gjson.Result
	for _, ctrl := range []string{"a", "b"} {
		postEndpoint := fmt.Sprintf("%s/storage-systems/%s/symbol/getFlashCacheStatistics?controller=%s&verboseErrorResponse=false",
			ep.Client.APIPath, systemID, ctrl)

		raw, postErr := ep.Client.Post(postEndpoint, []byte(`"`+cacheRef+`"`), headers)
		if postErr != nil {
			ep.Logger.Warn("failed to fetch SYMbol flash cache stats",
				slog.String("controller", ctrl), slogx.Err(postErr))
			continue
		}
		if len(raw) == 0 || raw[0].Get("returnCode").String() != "ok" {
			rc := ""
			if len(raw) > 0 {
				rc = raw[0].Get("returnCode").String()
			}
			ep.Logger.Warn("getFlashCacheStatistics returned non-ok",
				slog.String("controller", ctrl), slog.String("returnCode", rc))
			continue
		}

		// Wrap with synthetic controllerId ("a"/"b") so pollData and the Controller
		// plugin can create per-controller instances matching the REST path structure.
		wrapped := fmt.Sprintf(`{"controllerId":%q,"statistics":%s}`,
			ctrl, raw[0].Get("flashCacheStatistics").Raw)
		results = append(results, gjson.Parse(wrapped))
	}

	if len(results) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "getFlashCacheStatistics returned no data for either controller")
	}
	return results, nil
}
