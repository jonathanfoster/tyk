package gateway

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/config"
	"github.com/TykTechnologies/tyk/storage/kv"
	"github.com/TykTechnologies/tyk/test"
)

func TestTransformHeaders_EnabledForSpec(t *testing.T) {
	versionInfo := apidef.VersionInfo{
		GlobalHeaders: map[string]string{},
	}

	versions := map[string]apidef.VersionInfo{
		"Default": versionInfo,
	}

	th := TransformHeaders{BaseMiddleware: &BaseMiddleware{}}
	th.Spec = &APISpec{APIDefinition: &apidef.APIDefinition{}}
	th.Spec.VersionData.Versions = versions

	assert.False(t, th.EnabledForSpec())

	// version level add headers
	versionInfo.GlobalHeaders["a"] = "b"
	assert.True(t, versionInfo.GlobalHeadersEnabled())
	assert.True(t, th.EnabledForSpec())

	versionInfo.GlobalHeaders = nil
	versions["Default"] = versionInfo
	assert.False(t, th.EnabledForSpec())

	// endpoint level add headers
	versionInfo.UseExtendedPaths = true
	versionInfo.ExtendedPaths.TransformHeader = []apidef.HeaderInjectionMeta{{Disabled: false, DeleteHeaders: []string{"a"}}}
	versions["Default"] = versionInfo
	assert.True(t, th.EnabledForSpec())
}

func TestVersionInfo_GlobalHeadersEnabled(t *testing.T) {
	v := apidef.VersionInfo{
		GlobalHeaders:       map[string]string{},
		GlobalHeadersRemove: []string{},
	}

	assert.False(t, v.GlobalHeadersEnabled())

	// add headers
	v.GlobalHeaders["a"] = "b"
	assert.True(t, v.GlobalHeadersEnabled())
	v.GlobalHeadersDisabled = true
	assert.False(t, v.GlobalHeadersEnabled())

	// reset
	v.GlobalHeaders = map[string]string{}
	v.GlobalHeadersDisabled = false
	assert.False(t, v.GlobalHeadersEnabled())

	// remove headers
	v.GlobalHeadersRemove = []string{"a"}
	assert.True(t, v.GlobalHeadersEnabled())
	v.GlobalHeadersDisabled = true
	assert.False(t, v.GlobalHeadersEnabled())
}

func TestVersionInfo_HasEndpointReqHeader(t *testing.T) {
	v := apidef.VersionInfo{}

	assert.False(t, v.HasEndpointReqHeader())
	v.UseExtendedPaths = true
	assert.False(t, v.HasEndpointReqHeader())

	v.ExtendedPaths.TransformHeader = make([]apidef.HeaderInjectionMeta, 2)
	assert.False(t, v.HasEndpointReqHeader())

	v.ExtendedPaths.TransformHeader[0].Disabled = true
	v.ExtendedPaths.TransformHeader[0].AddHeaders = map[string]string{"a": "b"}
	assert.False(t, v.HasEndpointReqHeader())

	v.ExtendedPaths.TransformHeader[1].Disabled = false
	v.ExtendedPaths.TransformHeader[1].DeleteHeaders = []string{"a"}
	assert.True(t, v.HasEndpointReqHeader())
}

func TestHeaderInjectionMeta_Enabled(t *testing.T) {
	h := apidef.HeaderInjectionMeta{Disabled: true}
	assert.False(t, h.Enabled())

	h.Disabled = false
	assert.False(t, h.Enabled())

	h.AddHeaders = map[string]string{"a": "b"}
	assert.True(t, h.Enabled())

	h.AddHeaders = nil
	assert.False(t, h.Enabled())

	h.DeleteHeaders = []string{"a"}
	assert.True(t, h.Enabled())
}

func TestTransformHeadersReplaceSecrets(t *testing.T) {
	ts := StartTest(func(conf *config.Config) {
		conf.Secrets = map[string]string{
			"global_header": "conf::global_header::value",
			"path_header":   "conf::path_header::value",
		}
	})
	defer ts.Close()

	t.Setenv("TYK_SECRET_GLOBAL_HEADER", "env::global_header::value")
	t.Setenv("TYK_SECRET_PATH_HEADER", "env::path_header::value")

	ts.Gw.secretsManagerKVStore = kv.NewSecretsManagerWithClient(kv.NewDummySecretsManagerClient(map[string]string{
		"global_header": "secretsmanager::global_header::value",
		"path_header":   "secretsmanager::path_header::value",
	}))

	tests := []struct {
		name   string
		scheme string
	}{
		{name: "Config", scheme: "conf"},
		{name: "Env", scheme: "env"},
		{name: "SecretsManager", scheme: "secretsmanager"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headerCasedScheme := strings.ToUpper(tt.scheme[:1]) + tt.scheme[1:]
			ts.Gw.BuildAndLoadAPI(func(spec *APISpec) {
				spec.Proxy.ListenPath = "/"
				UpdateAPIVersion(spec, "v1", func(v *apidef.VersionInfo) {
					v.UseExtendedPaths = true
					v.ExtendedPaths.TransformHeader = []apidef.HeaderInjectionMeta{{
						AddHeaders: map[string]string{
							"Path-" + headerCasedScheme: "$secret_" + tt.scheme + ".path_header",
						},
						Method: http.MethodGet,
						Path:   "/path",
					}}
					v.GlobalHeaders = map[string]string{
						"Global-" + headerCasedScheme: "$secret_" + tt.scheme + ".global_header",
					}
				})
			})

			t.Run("GlobalHeaders", func(t *testing.T) {
				_, err := ts.Run(t, test.TestCase{
					Method:    http.MethodGet,
					Path:      "/",
					Code:      http.StatusOK,
					BodyMatch: `"Global-` + headerCasedScheme + `":"` + tt.scheme + `::global_header::value"`,
				})
				assert.NoError(t, err)
			})

			t.Run("PathHeaders", func(t *testing.T) {
				_, err := ts.Run(t, test.TestCase{
					Method:    http.MethodGet,
					Path:      "/path",
					Code:      http.StatusOK,
					BodyMatch: `"Path-` + headerCasedScheme + `":"` + tt.scheme + `::path_header::value"`,
				})
				assert.NoError(t, err)
			})
		})
	}
}
