package client

type Template struct {
	// TODO: Not sure the difference between these two
	Id        string
	Uuid      string
	NameLabel string
}

func (t Template) Compare(obj interface{}) bool {
	other := obj.(Template)
	if t.NameLabel == other.NameLabel {
		return true
	}
	return false
}

func (p Template) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	uuid := obj["uuid"].(string)
	name_label := obj["name_label"].(string)
	return Template{
		Id:        id,
		NameLabel: name_label,
		Uuid:      uuid,
	}
}

func (c *Client) GetTemplate(name string) (Template, error) {
	obj, err := c.FindFromGetAllObjects(Template{NameLabel: name})
	template := obj.(Template)

	if err != nil {
		return template, err
	}

	return template, nil
}
