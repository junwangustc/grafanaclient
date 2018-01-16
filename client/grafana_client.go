package client

import "github.com/junwangustc/grafanaclient/grafana"

type Client struct {
	Sess *grafana.Session
}

func NewClient(userName, password, url string) *Client {
	cl := &Client{}
	cl.Sess = grafana.NewSession(userName, password, url)
	return cl
}

func (c *Client) UpdateDashboard(j *Job, v *View) (panelUrl string, err error) {

	return "", nil
}

func (c *Client) DeletePanel(j *Job, v *View) error {
	return nil

}
func (c *Client) DeleteDashboard(j *Job) error {
	return nil
}
