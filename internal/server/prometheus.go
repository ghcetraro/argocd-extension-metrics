package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/common/config"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// ThresholdResponse represents the response format for a threshold.
type ThresholdResponse struct {
	Data  json.RawMessage `json:"data"`
	Key   string          `json:"key"`
	Name  string          `json:"name"`
	Color string          `json:"color"`
	Value string          `json:"value"`
	Unit  string          `json:"unit"`
}

// AggregatedResponse represents the final output response structure returned by execute function
type AggregatedResponse struct {
	Data       json.RawMessage     `json:"data"`
	Thresholds []ThresholdResponse `json:"thresholds,omitempty"`
}

type PrometheusProvider struct {
	logger   *zap.SugaredLogger
	provider v1.API
	config   *MetricsConfigProvider
}

func (pp *PrometheusProvider) getType() string {
	return PROMETHEUS_TYPE
}

// getDashboard returns the dashboard configuration for the specified application
func (pp *PrometheusProvider) getDashboard(ctx *gin.Context) {
	appName := ctx.Param("application")
	groupKind := ctx.Param("groupkind")
	app := pp.config.getApp(appName)
	if app == nil {
		ctx.JSON(http.StatusBadRequest, "Requested/Default Application not found")
		return
	}
	dash := app.getDashBoard(groupKind)

	if dash == nil {
		ctx.JSON(http.StatusBadRequest, "Requested/Default Dashboard not found")
		return
	}
	dash.ProviderType = pp.getType()
	ctx.JSON(http.StatusOK, dash)
}

func NewPrometheusProvider(prometheusConfig *MetricsConfigProvider, logger *zap.SugaredLogger) *PrometheusProvider {
	return &PrometheusProvider{config: prometheusConfig, logger: logger}
}

func (pp *PrometheusProvider) init() error {
	// Configurar HTTP client con TLS
	httpClientConfig := config.HTTPClientConfig{}
	
	// Si TLSConfig está configurado, usarlo; si no, usar InsecureSkipVerify para desarrollo
	if pp.config.Provider.TLSConfig.InsecureSkipVerify || 
		(pp.config.Provider.TLSConfig.CAFile == "" && 
		 pp.config.Provider.TLSConfig.CertFile == "" && 
		 pp.config.Provider.TLSConfig.KeyFile == "") {
		// No hay configuración TLS, usar InsecureSkipVerify para desarrollo
		httpClientConfig.TLSConfig = config.TLSConfig{
			InsecureSkipVerify: true,
		}
		pp.logger.Warnf("Using InsecureSkipVerify for Prometheus connection (development mode)")
	} else {
		// Usar la configuración TLS del provider
		httpClientConfig.TLSConfig = pp.config.Provider.TLSConfig
	}
	
	httpClient, err := config.NewClientFromConfig(httpClientConfig, "prometheus")
	if err != nil {
		pp.logger.Errorf("Error creating HTTP client: %v\n", err)
		return err
	}
	
	client, err := api.NewClient(api.Config{
		Address:      pp.config.Provider.Address,
		RoundTripper: httpClient.Transport,
	})
	if err != nil {
		pp.logger.Errorf("Error creating Prometheus client: %v\n", err)
		return err
	}
	pp.provider = v1.NewAPI(client)
	return nil
}

// executeGraphQuery executes a prometheus query and returns the result.
func executeGraphQuery(ctx *gin.Context, queryExpression string, env map[string][]string, duration time.Duration, pp *PrometheusProvider) (model.Value, v1.Warnings, error) {
	tmpl, err := template.New("query").Parse(queryExpression)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing query template: %s", err)
	}

	env1 := make(map[string]string)
	for k, v := range env {
		env1[k] = strings.Join(v, ",")
	}

	// Logging para diagnóstico
	pp.logger.Infof("Template queryExpression: %s", queryExpression)
	pp.logger.Infof("Template env (raw): %+v", env)
	pp.logger.Infof("Template env1 (processed): %+v", env1)
	if nameVal, ok := env1["name"]; ok {
		pp.logger.Infof("Template env1['name'] value: '%s'", nameVal)
	} else {
		pp.logger.Warnf("Template env1['name'] NOT FOUND in map")
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, env1)
	if err != nil {
		pp.logger.Errorf("Error executing template: %v", err)
		return nil, nil, fmt.Errorf("error executing template: %s", err)
	}

	strQuery := buf.String()
	pp.logger.Infof("Final query after template execution: %s", strQuery)
	r := v1.Range{
		Start: time.Now().Add(-duration),
		End:   time.Now(),
		Step:  time.Minute,
	}

	result, warnings, err := pp.provider.QueryRange(ctx, strQuery, r)

	if err != nil {
		return nil, warnings, fmt.Errorf("error querying prometheus: %s", err)
	}

	if len(warnings) > 0 {
		return result, warnings, fmt.Errorf("query warnings: %s", err)
	}

	return result, nil, nil
}

// execute handles the execution of a graph queryExpression and graph thresholds
func (pp *PrometheusProvider) execute(ctx *gin.Context) {
	app := ctx.Param("application")
	groupKind := ctx.Param("groupkind")
	rowName := ctx.Param("row")
	graphName := ctx.Param("graph")
	durationStr := ctx.Query("duration")
	if durationStr == "" {
		durationStr = "1h"
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid duration format :"+err.Error())
		return
	}

	env := ctx.Request.URL.Query()

	application := pp.config.getApp(app)
	if application == nil {
		ctx.JSON(http.StatusBadRequest, "Requested/Default Application not found")
		return
	}
	dashboard := application.getDashBoard(groupKind)
	if dashboard == nil {
		ctx.JSON(http.StatusBadRequest, "Requested/Default Dashboard not found")
		return
	}
	row := dashboard.getRow(rowName)
	if row == nil {
		ctx.JSON(http.StatusBadRequest, "Requested Row not found")
		return
	}
	graph := row.getGraph(graphName)
	if graph != nil {

		var data AggregatedResponse
		result, warnings, err := executeGraphQuery(ctx, graph.QueryExpression, env, duration, pp)

		if err != nil {
			pp.logger.Errorf("Error executing graph query: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(warnings) > 0 {
			warningMsg := fmt.Errorf("query warnings: %s", warnings)
			pp.logger.Warnf("Query warnings: %v", warnings)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": warningMsg.Error()})
			return
		}
		data.Data, err = json.Marshal(result)
		if err != nil {
			pp.logger.Errorf("Error marshaling data: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("error marshaling the data: %s", err)})
			return
		}
		var finalResultArr []ThresholdResponse
		if graph.Thresholds != nil {

			for _, threshold := range graph.Thresholds {
				var result model.Value
				var warnings v1.Warnings
				var err error

				//If threshold.value present, threshold.value gets executed else,threshold.queryExpression gets executed.
				if threshold.Value != "" {
					result, warnings, err = executeGraphQuery(ctx, threshold.Value, env, duration, pp)
				} else {
					result, warnings, err = executeGraphQuery(ctx, threshold.QueryExpression, env, duration, pp)
				}
				if err != nil {
					pp.logger.Errorf("Error executing threshold query: %v", err)
					ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if len(warnings) > 0 {
					warningMsg := fmt.Errorf("query warnings: %s", warnings)
					pp.logger.Warnf("Query warnings: %v", warnings)
					ctx.JSON(http.StatusBadRequest, gin.H{"error": warningMsg.Error()})
					return
				}
				var temp ThresholdResponse
				temp.Unit = threshold.Unit
				temp.Name = threshold.Name
				temp.Value = threshold.Value
				temp.Key = threshold.Key
				temp.Color = threshold.Color
				temp.Data, err = json.Marshal(result)
				if err != nil {
					pp.logger.Errorf("Error marshaling threshold response: %v", err)
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("error marshaling the threshold response: %s", err)})
					return
				}

				finalResultArr = append(finalResultArr, temp)
			}
		}
		data.Thresholds = finalResultArr

		ctx.JSON(http.StatusOK, data)
		return
	}
}
