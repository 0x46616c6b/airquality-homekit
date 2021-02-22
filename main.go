package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/0x46616c6b/airquality-homekit/sensors"
)

const (
	Manufacturer = "Raspberry"
	Model        = "Zero"
	SerialNumber = "AIRPIZERO1"
)

var (
	pin        = flag.String("pin", "36363636", "Pin which has to be entered on iOS client to pair with the accessory")
	metricsURL = flag.String("metrics-url", "http://localhost:9229/metrics", "URL of the airquality-exporter")

	bridge = accessory.NewBridge(accessory.Info{
		Name:         "Bridge",
		Manufacturer: Manufacturer,
		Model:        Model,
		SerialNumber: SerialNumber,
	})
	temperatureSensor = accessory.NewTemperatureSensor(accessory.Info{
		Name:         "Temperature",
		Manufacturer: Manufacturer,
		Model:        Model,
		SerialNumber: SerialNumber,
	}, 23.0, 0.0, 36, 0)
	airQualitySensor = sensors.NewAirQualitySensor(accessory.Info{
		Name:         "Air Quality",
		Manufacturer: Manufacturer,
		Model:        "Zero",
		SerialNumber: SerialNumber,
	})
	humiditySensor = sensors.NewHumiditySensor(accessory.Info{
		Name:         "Humidity",
		Manufacturer: Manufacturer,
		Model:        Model,
		SerialNumber: SerialNumber,
	})
)

func main() {
	flag.Parse()

	if os.Getenv("LOG_LEVEL") == "debug" {
		log.Debug.Enable()
	}

	config := hc.Config{Pin: *pin}
	t, err := hc.NewIPTransport(config, bridge.Accessory, temperatureSensor.Accessory, airQualitySensor.Accessory, humiditySensor.Accessory)
	if err != nil {
		log.Info.Panic(err)
	}

	go func() {
		var co2, tvoc float64

		for {
			time.Sleep(time.Second * 10)
			metrics, err := fetchMetrics()
			if err != nil {
				log.Info.Fatal(err)
			}

			for metric, value := range metrics {
				if len(value.Metric) > 0 {
					if metric == "airquality_temperature" {
						temp := value.Metric[0].Gauge.GetValue()
						temperatureSensor.TempSensor.CurrentTemperature.SetValue(temp)
					}
					if metric == "airquality_humidity" {
						hum := value.Metric[0].Gauge.GetValue()
						humiditySensor.HumiditySensor.CurrentRelativeHumidity.SetValue(hum)
					}
					if metric == "airquality_co2" {
						co2 = value.Metric[0].Gauge.GetValue()
					}
					if metric == "airquality_tvoc" {
						tvoc = value.Metric[0].Gauge.GetValue()
					}
				}
			}

			if tvoc > 8333 || co2 > 2500 {
				airQualitySensor.AirQualitySensor.AirQuality.SetValue(characteristic.AirQualityPoor)
			} else if tvoc > 3333 || co2 > 1500 {
				airQualitySensor.AirQualitySensor.AirQuality.SetValue(characteristic.AirQualityInferior)
			} else if tvoc > 1000 || co2 > 1000 {
				airQualitySensor.AirQualitySensor.AirQuality.SetValue(characteristic.AirQualityFair)
			} else if tvoc > 333 || co2 > 600 {
				airQualitySensor.AirQualitySensor.AirQuality.SetValue(characteristic.AirQualityGood)
			} else {
				airQualitySensor.AirQualitySensor.AirQuality.SetValue(characteristic.AirQualityExcellent)
			}
		}
	}()

	t.Start()
}

func fetchMetrics() (map[string]*dto.MetricFamily, error) {
	u, err := url.Parse(*metricsURL)
	if err != nil {
		return nil, err
	}

	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	var parser expfmt.TextParser
	metrics, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
