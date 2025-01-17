package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type WebhookNotify struct {
	Url        string
	Method     string
	HttpClient *http.Client
}

func NewWebhookNotify(url string, method string) *WebhookNotify {
	log.Printf("webhook url: %s\nmethod: %s\n", url, method)
	httpClient := &http.Client{}
	return &WebhookNotify{
		Url:        url,
		Method:     method,
		HttpClient: httpClient,
	}
}

func (w *WebhookNotify) Notify(summary, detail string) error {
	log.Printf("sending notify: %s to webhook\n", summary)

	var outBuffer *bytes.Buffer
	if w.Method == "POST" {
		payload := map[string]interface{}{
			"state":    "firing",
			"ruleName": "WatchdogAlert",
			"message":  fmt.Sprintf("%s\n%s", summary, detail),
			"ruleUrl":  "https://github.com/carv-protocol/kubernetes/blob/main/components/prometheus-alert/templates/watchdog.yaml",
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		outBuffer = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(w.Method, w.Url, outBuffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := w.HttpClient.Do(req)
	if err != nil {
		return err
	}
	log.Printf("webhook response: %s\n", res.Status)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Printf("webhook response body: %s\n", resBody)
	return nil
}

/*
{{ $var := .externalURL}}{{ range $k,$v:=.alerts }}{{if eq $v.status "resolved"}}**[Prometheus恢复信息]({{$v.generatorURL}})**
*[{{$v.labels.alertname}}]({{$var}})*
告警级别：{{$v.labels.level}}
开始时间：{{$v.startsAt}}
结束时间：{{$v.endsAt}}
故障主机IP：{{$v.labels.instance}}
**{{$v.annotations.description}}**{{else}}**[Prometheus告警信息]({{$v.generatorURL}})**
*[{{$v.labels.alertname}}]({{$var}})*
告警级别：{{$v.labels.level}}
开始时间：{{$v.startsAt}}
故障主机IP：{{$v.labels.instance}}
**{{$v.annotations.description}}**{{end}}{{ end }}
{{ $urimsg:=""}}{{ range $key,$value:=.commonLabels }}{{$urimsg =  print $urimsg $key "%3D%22" $value "%22%2C" }}{{end}}[*** 点我屏蔽该告警]({{$var}}/#/silences/new?filter=%7B{{SplitString $urimsg 0 -3}}%7D)


{
  "receiver": "webhook",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "WatchdogAlert"
      },
      "startsAt": "2018-08-03T09:52:26.739266876+02:00",
      "endsAt": "0001-01-01T00:00:00Z"
    }
  ],
  "groupLabels": {
    "alertname": "WatchdogAlert"
  },
  "commonLabels": {
    "alertname": "WatchdogAlert"
  },
  "commonAnnotations": {
    "summary": "{{ .summary }}",
    "description": "{{ .detail }}",
  },
  "version": "4",
  "groupKey": "{}:{alertname=\"WatchdogAlert\"}"
}
*/
