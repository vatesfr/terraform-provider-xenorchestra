package client

type Host struct {
	Id        string `json:"id"`
	NameLabel string `json:"name_label"`
}

func (h Host) Compare(obj interface{}) bool {
	otherHost := obj.(Host)
	if otherHost.NameLabel != "" && h.NameLabel == otherHost.NameLabel {
		return true
	}
	return false
}

func (c *Client) GetHostByName(nameLabel string) (hosts []Host, err error) {
	obj, err := c.FindFromGetAllObjects(Host{NameLabel: nameLabel})
	if err != nil {
		return
	}
	return obj.([]Host), nil
}
