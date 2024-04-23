package tmplview

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/naoina/toml"
)

var CatUiIdReg = regexp.MustCompile(`^[a-z0-9_]+$`)

type CatUiVar struct {
	Id      string
	Label   string
	Comment string `toml:",omitempty"`
}
type CatUiConfig struct {
	Url string
	Var []CatUiVar
}
type CatUiParam struct {
	CatUiConfig
	Name string
}

func (tmpv *TmplView) load_cat_ui_cfg(api_name string) (*CatUiParam, error) {
	cfg_file, jerr := tmpv.CatUiConfigPath.Join(api_name + "." + tmpv.CatUiConfigExt)
    if jerr != nil {
        return nil, jerr
    }

	cfg_f, err := cfg_file.Open(tmpv.SystemFS)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, new_err("Cat UI Config file open error: %s", cfg_file)
		}
		return nil, err
	}
	defer cfg_f.Close()

	text, err := io.ReadAll(cfg_f)
	if err != nil {
		return nil, new_err("Cat UI Config file read error: %s", cfg_file)
	}

	cfg := &CatUiParam{}
	if err := toml.Unmarshal(text, &cfg.CatUiConfig); err != nil {
		return nil, new_err("Config file parse error: %s", cfg_file)
	}
	cfg.Name = strings.ReplaceAll(api_name, "/", "-")

	for _, e := range cfg.Var {
		if !CatUiIdReg.MatchString(e.Id) {
			return nil, new_err("Cat UI Var.Id parameter error: %s", cfg_file)
		}
	}

	return cfg, nil
}

func (tmpv *TmplView) WriteTestCatUi(w io.Writer) error {
	if tmpv.CatUiTmplName == "" {
		return nil
	}

	if err := tmpv.writeTestCatUi1(w); err != nil {
		return err
	}
	return tmpv.writeTestCatUi2(w)
}

func (tmpv *TmplView) writeTestCatUi1(w io.Writer) error {
	cfg := CatUiConfig{
		Url: "/",
		Var: []CatUiVar{{Id: "test", Label: "test"}},
	}
	param := &CatUiParam{
		CatUiConfig: cfg,
		Name:        "test",
	}
	return tmpv.OriginTmpl.ExecuteTemplate(w, tmpv.CatUiTmplName, param)
}

func (tmpv *TmplView) writeTestCatUi2(w io.Writer) error {
	cfg := CatUiConfig{
		Url: "/",
		Var: []CatUiVar{},
	}
	param := &CatUiParam{
		CatUiConfig: cfg,
		Name:        "test",
	}
	return tmpv.OriginTmpl.ExecuteTemplate(w, tmpv.CatUiTmplName, param)
}

func (tmpv *TmplView) CatUi(api_name string) string {
	if tmpv.CatUiConfigPath.IsZero() {
		return ""
	}
	if tmpv.CatUiTmplName == "" {
		return ""
	}

	catui_cfg, err := tmpv.load_cat_ui_cfg(api_name)
	if err != nil {
		tmpv.Warn("%s", err.Error())
		return ""
	}

	var uibuf bytes.Buffer
	if err := tmpv.OriginTmpl.ExecuteTemplate(
		&uibuf, tmpv.CatUiTmplName, catui_cfg); err != nil {
		tmpv.Warn("Cat UI template error: %s", err)
		return ""
	}
	return uibuf.String()
}
