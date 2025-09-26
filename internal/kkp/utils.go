package kkp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func normalizeBase(endpoint string) (*url.URL, error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Host == "" && u.Path != "" {
		// allow "kkp.example.com[/api]" without scheme
		u.Host = u.Path
		u.Path = ""
	}
	u.Path = stripTrailingAPI(u.Path) // <-- key change
	return u, nil
}

// If path ends with "/api", drop that suffix. Keep any prefix (e.g., "/kkp").
func stripTrailingAPI(p string) string {
	clean := path.Clean("/" + strings.TrimSpace(p))
	if clean == "/" {
		return ""
	}
	if strings.HasSuffix(clean, "/api") {
		return strings.TrimSuffix(clean, "/api")
	}
	return clean
}

func buildTLSConfig(insecure bool, caFile string) (*tls.Config, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if caFile != "" {
		pem, err := os.ReadFile(caFile) // #nosec G304 -- Reading CA file is legitimate in TLS config
		if err != nil {
			return nil, fmt.Errorf("read CAFile: %w", err)
		}
		if ok := rootCAs.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to append certs from %s", caFile)
		}
	}
	return &tls.Config{
		InsecureSkipVerify: insecure, //nolint:gosec // user-configurable for dev/self-signed endpoints
		RootCAs:            rootCAs,
		MinVersion:         tls.VersionTLS12,
	}, nil
}

func newHTTPClient(timeout time.Duration, tlsCfg *tls.Config) *http.Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsCfg,
		},
		Timeout: timeout,
	}
}

func defaultUA(ua string) string {
	if strings.TrimSpace(ua) != "" {
		return ua
	}
	return "terraform-provider-kkp"
}

// VariablesToJSON converts variables interface to JSON string
func VariablesToJSON(variables interface{}) (string, error) {
	if variables == nil {
		return "{}", nil
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// JSONToVariables converts JSON string to variables interface
func JSONToVariables(jsonStr string) (interface{}, error) {
	if strings.TrimSpace(jsonStr) == "" || jsonStr == "{}" {
		return nil, nil
	}

	var variables interface{}
	if err := json.Unmarshal([]byte(jsonStr), &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

// TrimmedStringValue extracts and trims string value from Terraform attribute
func TrimmedStringValue(attr tftypes.String) string {
	return strings.TrimSpace(attr.ValueString())
}

// IsAttributeSet checks if a Terraform attribute is set (not null and not unknown)
func IsAttributeSet(attr interface{}) bool {
	switch v := attr.(type) {
	case tftypes.String:
		return !v.IsNull() && !v.IsUnknown()
	case tftypes.Bool:
		return !v.IsNull() && !v.IsUnknown()
	case tftypes.Int64:
		return !v.IsNull() && !v.IsUnknown()
	default:
		return false
	}
}

// MergeInt64 copies src into dst when dst is null or unknown.
func MergeInt64(dst *tftypes.Int64, src tftypes.Int64) {
	if dst == nil {
		return
	}
	if dst.IsNull() || dst.IsUnknown() {
		*dst = src
	}
}

// ---------- Plan Modifiers ----------

// Int64RequiresReplaceModifier forces replacement when an int64 attribute changes to a different known value.
type Int64RequiresReplaceModifier struct{}

func (Int64RequiresReplaceModifier) Description(context.Context) string {
	return "Requires replacement when the attribute changes to a different known value."
}

func (Int64RequiresReplaceModifier) MarkdownDescription(ctx context.Context) string {
	return Int64RequiresReplaceModifier{}.Description(ctx)
}

func (Int64RequiresReplaceModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	if req.StateValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.ValueInt64() != req.StateValue.ValueInt64() {
		resp.RequiresReplace = true
	}
}

// MergeString copies src into dst when dst is null or unknown.
func MergeString(dst *tftypes.String, src tftypes.String) {
	if dst == nil {
		return
	}
	if dst.IsNull() || dst.IsUnknown() {
		*dst = src
	}
}

// MergeBool copies src into dst when dst is null or unknown.
func MergeBool(dst *tftypes.Bool, src tftypes.Bool) {
	if dst == nil {
		return
	}
	if dst.IsNull() || dst.IsUnknown() {
		*dst = src
	}
}

// ExtractIDs extracts and validates ID and ClusterID from state
func ExtractIDs(id, clusterID tftypes.String) (string, string, error) {
	idStr := TrimmedStringValue(id)
	clusterIDStr := TrimmedStringValue(clusterID)

	if idStr == "" || clusterIDStr == "" {
		return "", "", fmt.Errorf("missing required identifiers")
	}

	return idStr, clusterIDStr, nil
}

// SafeInt32 safely converts int64 to int32, returning an error if overflow would occur
func SafeInt32(value int64) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("value %d would overflow int32 (range: %d to %d)", value, math.MinInt32, math.MaxInt32)
	}
	return int32(value), nil
}

// K8sVersionPattern validates Kubernetes version format (e.g. "1.28.5" or "1.28")
var K8sVersionPattern = regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`)

// Add more patterns as they emerge across resources
// Example: ResourceNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ---------- Label Conversion Helpers ----------

// ConvertLabelsToTerraform converts KKP labels (map[string]string) to Terraform MapValue
func ConvertLabelsToTerraform(labels map[string]string) tftypes.Map {
	labelMap := map[string]attr.Value{}
	for key, value := range labels {
		labelMap[key] = tftypes.StringValue(value)
	}
	return tftypes.MapValueMust(tftypes.StringType, labelMap)
}

// ---------- Plan Execution Helpers ----------

// ExecutePlan provides common plan execution flow for any plan type
func ExecutePlan[T PlanValidator](plan T) error {
	plan.SetDefaults()
	return plan.Validate()
}

// ExecuteToModel wraps the common pattern of SetDefaults + Validate before model creation
func ExecuteToModel[T PlanValidator, R any](plan T, buildFunc func() (R, error)) (R, error) {
	plan.SetDefaults()
	if err := plan.Validate(); err != nil {
		var zero R
		return zero, err
	}
	return buildFunc()
}
