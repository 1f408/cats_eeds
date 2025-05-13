package md2html

import (
	_ "embed"

	"github.com/naoina/toml"

	"github.com/1f408/cats_eeds/md2html/alerts"
)

//go:embed "alert_title_mapping.conf"
var defaultAlertTitleMapping []byte

type AlertTitleMapping alerts.TitleHtmlMapping

func (ppm *AlertTitleMapping) Initialize() {
	type Raw AlertTitleMapping
	if err := toml.Unmarshal(defaultAlertTitleMapping, (*Raw)(ppm)); err != nil {
		panic("bad default alert title mappiing file: " + err.Error())
	}
}

func (_ *AlertTitleMapping) MakeNew() *AlertTitleMapping {
	thm := &AlertTitleMapping{}
	thm.Initialize()

	return thm
}

func (thm *AlertTitleMapping) UnmarshalTOML(decode func(interface{}) error) error {
	if thm == nil {
		thm.Initialize()
	}

	type Raw AlertTitleMapping
	return decode((*Raw)(thm))
}
