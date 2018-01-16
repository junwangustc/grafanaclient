package grafana

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"time"
)

const timeout = 5

var protocolRegexp = regexp.MustCompile(`^https://`)

// GrafanaError is a error structure to handle error messages in this library
type GrafanaError struct {
	Code        int
	Description string
}

// A GrafanaMessage contains the json error message received when http request failed
type GrafanaMessage struct {
	Message string `json:"message"`
}

// Error generate a text error message.
// If Code is zero, we know it's not a http error.
func (h GrafanaError) Error() string {
	if h.Code != 0 {
		return fmt.Sprintf("HTTP %d: %s", h.Code, h.Description)
	}
	return fmt.Sprintf("ERROR: %s", h.Description)
}

type DashboardResult struct {
	Meta  Meta      `json:"meta"`
	Model Dashboard `json:"model"`
}

// A Meta contains a Dashboard metadata.
type Meta struct {
	Created    string `json:"created"`
	Expires    string `json:"expires"`
	IsHome     bool   `json:"isHome"`
	IsSnapshot bool   `json:"isSnapshot"`
	IsStarred  bool   `json:"isStarred"`
	Slug       string `json:"slug"`
}

// A Dashboard contains the Dashboard structure.
type Dashboard struct {
	Editable      bool          `json:"editable"`
	GnetID        interface{}   `json:"gnetId"`
	GraphTooltip  int           `json:"graphTooltip"`
	HideControls  bool          `json:"hideControls"`
	ID            int           `json:"id"`
	Links         []interface{} `json:"links"`
	Rows          []Row         `json:"rows"`
	SchemaVersion int           `json:"schemaVersion"`
	Style         string        `json:"style"`
	Tags          []interface{} `json:"tags"`
	Templating    Templating    `json:"templating"`
	Time          Time          `json:"time"`
	Timepicker    Timepicker    `json:"timepicker"`
	Timezone      string        `json:"timezone"`
	Title         string        `json:"title"`
	Version       int           `json:"version"`
}

type Templating struct {
	List []Template `json:"list"`
}

func GetDefaultTemplating(tagNames []string, measurementName, datasource string) Templating {
	tp := Templating{}
	tp.List = GetDefaultTemplates(tagNames, measurementName, datasource)
	return tp
}

type Time struct {
	From string `json:"from"`
	To   string `json:"to"`
}
type Timepicker struct {
	RefreshIntervals []string `json:"refresh_intervals"`
	TimeOptions      []string `json:"time_options"`
}

func GetDefaultDashBoard(dashboardTitle string) *Dashboard {
	db := &Dashboard{}
	db.Editable = true
	db.GnetID = nil
	db.GraphTooltip = 0
	db.HideControls = false
	db.Links = make([]interface{}, 0)
	db.Rows = make([]Row, 0)
	db.SchemaVersion = 14
	db.Style = "dark"
	db.Tags = make([]interface{}, 0)
	db.Templating = Templating{List: nil}
	db.Time = Time{From: "now-6h", To: "now"}
	db.Timepicker = Timepicker{RefreshIntervals: []string{"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"}, TimeOptions: []string{"5m", "15m", "1h", "6h", "12h", "24h", "2d", "4d", "7d", "30d"}}
	db.Timezone = "browser"
	db.Title = dashboardTitle
	db.Version = 1
	return db
}

type Row struct {
	Collapse        bool        `json:"collapse"`
	Height          string      `json:"height"`
	Panels          []Panel     `json:"panels"`
	Repeat          interface{} `json:"repeat"`
	RepeatIteration interface{} `json:"repeatIteration"`
	RepeatRowID     interface{} `json:"repeatRowId"`
	ShowTitle       bool        `json:"showTitle"`
	Title           string      `json:"title"`
	TitleSize       string      `json:"titleSize"`
}

//每次只创建一行，一行只创建一个面板
func GetDefaultRow(panelTitle string, influxql string) Row {
	row := Row{}
	row.Collapse = false
	row.Height = "250px"
	row.Panels = make([]Panel, 0)
	row.Repeat = nil
	row.RepeatIteration = nil
	row.RepeatRowID = nil
	row.ShowTitle = false
	row.Title = ""
	row.TitleSize = "h6"

	panel := GetDefaultPanel(panelTitle, influxql)
	row.Panels = append(row.Panels, panel)
	return row
}

type Panel struct {
	AliasColors     struct{}      `json:"aliasColors"`
	Bars            bool          `json:"bars"`
	Datasource      interface{}   `json:"datasource"`
	Fill            int           `json:"fill"`
	ID              int           `json:"id"`
	Legend          Legend        `json:"legend"`
	Lines           bool          `json:"lines"`
	Linewidth       int           `json:"linewidth"`
	Links           []interface{} `json:"links"`
	NullPointMode   string        `json:"nullPointMode"`
	Percentage      bool          `json:"percentage"`
	Pointradius     int           `json:"pointradius"`
	Points          bool          `json:"points"`
	Renderer        string        `json:"renderer"`
	SeriesOverrides []interface{} `json:"seriesOverrides"`
	Span            int           `json:"span"`
	Stack           bool          `json:"stack"`
	SteppedLine     bool          `json:"steppedLine"`
	Targets         []Target      `json:"targets"`
	Thresholds      []interface{} `json:"thresholds"`
	TimeFrom        interface{}   `json:"timeFrom"`
	TimeShift       interface{}   `json:"timeShift"`
	Title           string        `json:"title"`
	Tooltip         Tooltip       `json:"tooltip"`
	Type            string        `json:"type"`
	Xaxis           Xaxis         `json:"xaxis"`
	Yaxes           []Yaxes       `json:"yaxes"`
}

func GetDefaultPanel(title string, influxql string) Panel {
	panel := Panel{}
	panel.Bars = false
	panel.Datasource = nil
	panel.Fill = 1
	panel.Legend = GetDefaultLegend()
	panel.Lines = true
	panel.Linewidth = 1
	panel.Links = make([]interface{}, 0)
	panel.NullPointMode = "null"
	panel.Percentage = false
	panel.Pointradius = 5
	panel.Points = false
	panel.Renderer = "flot"
	panel.SeriesOverrides = make([]interface{}, 0)
	panel.Span = 12
	panel.Stack = false
	panel.SteppedLine = false
	panel.Targets = GetDefaultTargets(influxql)
	panel.Thresholds = make([]interface{}, 0)
	panel.TimeFrom = nil
	panel.TimeShift = nil
	panel.Title = title
	panel.Tooltip = GetDefaultToolTip()
	panel.Type = "graph"
	panel.Xaxis = GetDefaultXaxis()
	panel.Yaxes = GetDefaultYaxes()
	return panel
}

type Legend struct {
	Avg     bool `json:"avg"`
	Current bool `json:"current"`
	Max     bool `json:"max"`
	Min     bool `json:"min"`
	Show    bool `json:"show"`
	Total   bool `json:"total"`
	Values  bool `json:"values"`
}

func GetDefaultLegend() Legend {
	return Legend{
		Avg:     false,
		Current: false,
		Max:     false,
		Min:     false,
		Show:    true,
		Total:   false,
		Values:  false,
	}
}

type Target struct {
	DsType  string `json:"dsType"`
	GroupBy []struct {
		Params []string `json:"params"`
		Type   string   `json:"type"`
	} `json:"groupBy"`
	Measurement  string `json:"measurement"`
	Policy       string `json:"policy"`
	Query        string `json:"query"`
	RawQuery     bool   `json:"rawQuery"`
	RefID        string `json:"refId"`
	ResultFormat string `json:"resultFormat"`
	Select       [][]struct {
		Params []string `json:"params"`
		Type   string   `json:"type"`
	} `json:"select"`
	Tags []interface{} `json:"tags"`
}

func GetDefaultTargets(influxql string) []Target {
	res := make([]Target, 0)
	targets := Target{}
	targets.DsType = "influxdb"
	targets.Policy = "default"
	targets.Query = influxql
	targets.RawQuery = true
	targets.RefID = "A"
	targets.ResultFormat = "time_series"
	res = append(res, targets)
	return res
}

type Tooltip struct {
	Shared    bool   `json:"shared"`
	Sort      int    `json:"sort"`
	ValueType string `json:"value_type"`
}

func GetDefaultToolTip() Tooltip {
	return Tooltip{
		Shared:    true,
		Sort:      0,
		ValueType: "individual",
	}
}

type Xaxis struct {
	Mode   string        `json:"mode"`
	Name   interface{}   `json:"name"`
	Show   bool          `json:"show"`
	Values []interface{} `json:"values"`
}

func GetDefaultXaxis() Xaxis {
	return Xaxis{
		Mode:   "time",
		Name:   nil,
		Show:   true,
		Values: make([]interface{}, 0),
	}
}

type Yaxes struct {
	Format  string      `json:"format"`
	Label   interface{} `json:"label"`
	LogBase int         `json:"logBase"`
	Max     interface{} `json:"max"`
	Min     interface{} `json:"min"`
	Show    bool        `json:"show"`
}

func GetDefaultYaxes() []Yaxes {
	yaxes := make([]Yaxes, 0)
	yax := Yaxes{
		Format:  "short",
		Label:   nil,
		LogBase: 1,
		Max:     nil,
		Min:     nil,
		Show:    true,
	}
	yaxes = append(yaxes, yax)
	yaxes = append(yaxes, yax)
	return yaxes
}

type UserInfo struct {
	User     string `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Template struct {
	Current struct {
		Tags  []interface{} `json:"tags"`
		Text  string        `json:"text"`
		Value interface{}   `json:"value"`
	} `json:"current,omitempty"`
	Datasource string `json:"datasource"`
	Hide       int    `json:"hide"`
	IncludeAll bool   `json:"includeAll"`
	Label      string `json:"label"`
	Multi      bool   `json:"multi"`
	Name       string `json:"name"`
	Options    []struct {
		Selected bool   `json:"selected"`
		Text     string `json:"text"`
		Value    string `json:"value"`
	} `json:"options,omitempty"`
	Query   string `json:"query"`
	Refresh int    `json:"refresh"`
	Regex   string `json:"regex"`
	Sort    int    `json:"sort"`
	Type    string `json:"type"`
	UseTags bool   `json:"useTags"`
}

func GetDefaultTemplates(qls []string, measurementName, datasource string) []Template {
	tpls := make([]Template, 0)
	for _, ql := range qls {
		tpls = append(tpls, GetDefaultTemplate(ql, measurementName, datasource))
	}
	return tpls
}
func GetDefaultTemplate(tagName, measurementName, datasource string) Template {
	tpl := Template{}
	tpl.Datasource = datasource
	tpl.Hide = 0
	tpl.IncludeAll = false
	tpl.Label = tagName
	tpl.Multi = true
	tpl.Name = tagName
	tpl.Query = "SHOW TAG VALUES FROM \"" + measurementName + "\" WITH  KEY = \"" + tagName + "\""
	tpl.Refresh = 1
	tpl.Sort = 0
	tpl.Type = "query"
	tpl.UseTags = false
	return tpl
}

type Session struct {
	client   *http.Client
	User     string
	Password string
	url      string
}

func NewSession(user string, password string, url string) *Session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{Jar: jar, Timeout: time.Second * timeout}
	if protocolRegexp.MatchString(url) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	return &Session{client: &client, User: user, Password: password, url: url}
}

func (s *Session) Login() (err error) {
	reqURL := s.url + "/login"
	loginInfo := UserInfo{User: s.User, Password: s.Password}
	jsonStr, _ := json.Marshal(loginInfo)
	_, err = s.httpRequest("POST", reqURL, bytes.NewBuffer(jsonStr))

	return

}
func (s *Session) httpRequest(method string, url string, body io.Reader) (result io.Reader, err error) {
	request, err := http.NewRequest(method, url, body)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return result, GrafanaError{0, "Unable to perform the http request"}
	}
	//    defer response.Body.Close()
	if response.StatusCode != 200 {
		dec := json.NewDecoder(response.Body)
		var gMess GrafanaMessage
		dec.Decode(&gMess)

		return result, GrafanaError{response.StatusCode, gMess.Message}
	}
	result = response.Body
	return
}

func (s *Session) Logout() {

}
func (s *Session) CreateDashboard(dashboardName string) Dashboard {
	db := GetDefaultDashBoard(dashboardName)
	return *db
}
func (s *Session) AddRowPanel(db Dashboard, panelTitle, influxql string) Dashboard {
	db.Rows = append(db.Rows, GetDefaultRow(panelTitle, influxql))
	return db
}

func (s *Session) AddTemplating(db Dashboard, tagNames []string, measurementName, datasource string) Dashboard {
	db.Templating = GetDefaultTemplating(tagNames, measurementName, datasource)
	return db
}

type DashboardUploader struct {
	Dashboard Dashboard `json:"dashboard"`
	Overwrite bool      `json:"overwrite"`
}

func (s *Session) UpdateDashboard(db Dashboard, overwrite bool) (err error) {
	reqURL := s.url + "/api/dashboards/db"
	var content DashboardUploader
	content.Dashboard = db
	content.Overwrite = overwrite
	jsonStr, _ := json.Marshal(content)
	_, err = s.httpRequest("POST", reqURL, bytes.NewBuffer(jsonStr))
	return
}
func (s *Session) GetDashboard(name string) (dashboard DashboardResult, err error) {
	reqURL := s.url + "/api/dashboards/db/" + name
	body, err := s.httpRequest("GET", reqURL, nil)
	if err != nil {
		return
	}
	dec := json.NewDecoder(body)
	err = dec.Decode(&dashboard)
	return
}
func (s *Session) DeleteDashBoard(dashBoardName string) (err error) {
	dashRes, err := s.GetDashboard(dashBoardName)
	if err != nil {
		return
	}
	slug := dashRes.Meta.Slug
	reqURL := fmt.Sprintf("%s/api/dashboards/db/%s", s.url, slug)
	_, err = s.httpRequest("DELETE", reqURL, nil)
	return

	return nil
}
func (s *Session) CreateDataSource() {

}
func (s *Session) DeleteDataSource() {

}
