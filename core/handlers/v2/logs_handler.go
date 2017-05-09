package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/SpectoLabs/hoverfly/core/handlers"
	"github.com/codegangsta/negroni"
	"github.com/go-zoo/bone"
)

type HoverflyLogs interface {
	GetLogs(limit int) []*logrus.Entry
}

type LogsHandler struct {
	Hoverfly HoverflyLogs
}

var DefaultLimit = 500

func (this *LogsHandler) RegisterRoutes(mux *bone.Mux, am *handlers.AuthHandler) {
	mux.Get("/api/v2/logs", negroni.New(
		negroni.HandlerFunc(am.RequireTokenAuthentication),
		negroni.HandlerFunc(this.Get),
	))

	mux.Get("/api/v2/ws/logs", http.HandlerFunc(this.GetWS))
}

func (this *LogsHandler) Get(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var logs []*logrus.Entry

	queryParams := req.URL.Query()
	limitQuery, _ := strconv.Atoi(queryParams.Get("limit"))
	if limitQuery == 0 {
		limitQuery = DefaultLimit
	}

	logs = this.Hoverfly.GetLogs(limitQuery)

	if strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
		handlers.WriteResponse(w, []byte(logsToPlainText(logs)))
	} else {
		bytes, _ := json.Marshal(logsToLogsView(logs))
		handlers.WriteResponse(w, bytes)
	}
}

func logsToLogsView(logs []*logrus.Entry) LogsView {
	var logInterfaces []map[string]interface{}
	for _, entry := range logs {
		data := make(map[string]interface{}, len(entry.Data)+3)

		for k, v := range entry.Data {
			data[k] = v
		}

		data["time"] = entry.Time.Format(logrus.DefaultTimestampFormat)
		data["msg"] = entry.Message
		data["level"] = entry.Level.String()

		logInterfaces = append(logInterfaces, data)
	}

	return LogsView{
		Logs: logInterfaces,
	}
}

func logsToPlainText(logs []*logrus.Entry) string {

	var buffer bytes.Buffer
	for _, entry := range logs {
		entry.Logger = logrus.New()
		log, err := entry.String()
		if err == nil {
			buffer.WriteString(log)
		}
	}

	return buffer.String()
}

func (this *LogsHandler) GetWS(w http.ResponseWriter, r *http.Request) {

	var position int
	var previousLength int

	handlers.NewWebsocket(func() ([]byte, error) {
		logs := this.Hoverfly.GetLogs(500)
		logsLength := len(logs)

		position = position + logsLength - previousLength

		if position != 0 {
			position--
			logToPrint := logs[position]
			previousLength = logsLength
			logsView := logsToLogsView([]*logrus.Entry{logToPrint})
			return json.Marshal(logsView.Logs[0])
		}

		return nil, errors.New("No update needed")
	}, w, r)
}
