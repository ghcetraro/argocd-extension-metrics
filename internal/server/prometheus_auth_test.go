package server

import (
	"os"
	"testing"

	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNormalizePrometheusAddress(t *testing.T) {
	assert.Equal(t, "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom",
		normalizePrometheusAddress("https://prometheus-prod-13-prod-us-east-0.grafana.net"))
	assert.Equal(t, "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom",
		normalizePrometheusAddress("https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom"))
	assert.Equal(t, "https://prometheus-prod-13-prod-us-east-0.grafana.net/api/prom",
		normalizePrometheusAddress("https://prometheus-prod-13-prod-us-east-0.grafana.net/"))
}

func TestApplyPrometheusAuth_NoAuth(t *testing.T) {
	cfg := &config.HTTPClientConfig{}
	p := &provider{}
	err := applyPrometheusAuth(cfg, p, zap.NewNop().Sugar())
	require.NoError(t, err)
	assert.Nil(t, cfg.BasicAuth)
	assert.Empty(t, cfg.BearerToken)
	assert.Empty(t, cfg.BearerTokenFile)
}

func TestApplyPrometheusAuth_BasicFromConfig(t *testing.T) {
	cfg := &config.HTTPClientConfig{}
	p := &provider{
		BasicAuth: &config.BasicAuth{
			Username: "prom",
			Password: "secret",
		},
	}
	err := applyPrometheusAuth(cfg, p, zap.NewNop().Sugar())
	require.NoError(t, err)
	require.NotNil(t, cfg.BasicAuth)
	assert.Equal(t, "prom", cfg.BasicAuth.Username)
	assert.Equal(t, config.Secret("secret"), cfg.BasicAuth.Password)
}

func TestApplyPrometheusAuth_BasicFromEnvOverridesConfig(t *testing.T) {
	t.Setenv("PROMETHEUS_USERNAME", "from-env")
	t.Setenv("PROMETHEUS_PASSWORD", "env-pass")
	cfg := &config.HTTPClientConfig{}
	p := &provider{
		BasicAuth: &config.BasicAuth{
			Username: "from-config",
			Password: "config-pass",
		},
	}
	err := applyPrometheusAuth(cfg, p, zap.NewNop().Sugar())
	require.NoError(t, err)
	require.NotNil(t, cfg.BasicAuth)
	assert.Equal(t, "from-env", cfg.BasicAuth.Username)
	assert.Equal(t, config.Secret("env-pass"), cfg.BasicAuth.Password)
}

func TestApplyPrometheusAuth_BearerFromEnv(t *testing.T) {
	t.Setenv("PROMETHEUS_BEARER_TOKEN", "tok123")
	cfg := &config.HTTPClientConfig{}
	p := &provider{}
	err := applyPrometheusAuth(cfg, p, zap.NewNop().Sugar())
	require.NoError(t, err)
	assert.Equal(t, config.Secret("tok123"), cfg.BearerToken)
	assert.Nil(t, cfg.BasicAuth)
}

func TestApplyPrometheusAuth_ConflictBasicAndBearer(t *testing.T) {
	t.Setenv("PROMETHEUS_BEARER_TOKEN", "tok")
	t.Setenv("PROMETHEUS_USERNAME", "u")
	t.Setenv("PROMETHEUS_PASSWORD", "p")
	cfg := &config.HTTPClientConfig{}
	err := applyPrometheusAuth(cfg, &provider{}, zap.NewNop().Sugar())
	require.Error(t, err)
}

func TestApplyPrometheusAuth_BasicMissingPassword(t *testing.T) {
	_ = os.Unsetenv("PROMETHEUS_PASSWORD")
	_ = os.Unsetenv("PROMETHEUS_PASSWORD_FILE")
	cfg := &config.HTTPClientConfig{}
	p := &provider{
		BasicAuth: &config.BasicAuth{Username: "only-user"},
	}
	err := applyPrometheusAuth(cfg, p, zap.NewNop().Sugar())
	require.Error(t, err)
}
