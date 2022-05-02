package supermicro

import "encoding/xml"

type BiosCfg struct {
	XMLName xml.Name `xml:"BiosCfg"`
	Text    string   `xml:",chardata"`
	Menu    []*Menu  `xml:"Menu,omitempty"`
}

type Menu struct {
	Name    string     `xml:"name,attr"`
	Setting []*Setting `xml:"Setting,omitempty"`
	Menu    []*Menu    `xml:"Menu,omitempty"`
}

type Setting struct {
	Name           string `xml:"name,attr"`
	Type           string `xml:"type,attr"`
	SelectedOption string `xml:"selectedOption,attr,omitempty"`
	CheckedStatus  string `xml:"checkedStatus,attr,omitempty"`
}
