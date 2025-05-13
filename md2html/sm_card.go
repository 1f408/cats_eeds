package md2html

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/l4go/rpath"
)

type SmCardConfig struct {
	Enabled    bool   `toml:",omitempty"`
	SiteTopUrl string `toml:",omitempty"`

	Default struct {
		Image         string `toml:",omitempty"`
		UseXcomLarge  bool   `toml:",omitempty"`
		Description   string `toml:",omitempty"`
		SiteName      string `toml:",omitempty"`
		SiteXcomId    string `toml:",omitempty"`
		CreatorXcomId string `toml:",omitempty"`
	}
}

type SmCardParam struct {
	Description string `yaml:"description" toml:"description" json:"description"`

	Image        string `yaml:"image,omitempty" toml:"image,omitempty" json:"image,omitempty"`
	UseXcomLarge bool   `yaml:"is_large,omitempty" toml:"card_type,omitempty" json:"card_type,omitempty"`
	Alt          string `yaml:"alt,omitempty" toml:"alt,omitempty" json:"alt,omitempty"`

	Title         string `yaml:"title,omitempty" toml:"title,omitempty" json:"title,omitempty"`
	SiteName      string `yaml:"site_name,omitempty" toml:"site_name,omitempty" json:"site_name,omitempty"`
	SiteUrl       string `yaml:"site_url,omitempty" toml:"site_url,omitempty" json:"site_url,omitempty"`
	SiteXcomId    string `yaml:"site_xcom_id,omitempty" toml:"site_xcom_id,omitempty" json:"site_xcom_id,omitempty"`
	CreatorXcomId string `yaml:"creator_xcom_id,omitempty" toml:"creator_xcom_id,omitempty" json:"creator_xcom_id,omitempty"`
}

func (scp *SmCardParam) Fix(scc *SmCardConfig, abs_cur string) {
	if !scc.Enabled {
		*scp = SmCardParam{}
		return
	}

	if scc.SiteTopUrl == "" {
		*scp = SmCardParam{}
		return
	}

	if scp.Description == "" {
		scp.Description = scc.Default.Description
	}
	if scp.Description == "" {
		*scp = SmCardParam{}
		return
	}

	if scp.Image == "" {
		scp.Image = scc.Default.Image
		scp.UseXcomLarge = scc.Default.UseXcomLarge
	}
	if scp.SiteName == "" {
		scp.SiteName = scc.Default.SiteName
	}
	if scp.SiteXcomId == "" {
		scp.SiteXcomId = scc.Default.SiteXcomId
	}
	if scp.CreatorXcomId == "" {
		scp.CreatorXcomId = scc.Default.CreatorXcomId
	}

	scp.SiteUrl = newSiteUrl(scc.SiteTopUrl, abs_cur, scp.SiteUrl)
	if scp.Image != "" {
		scp.Image = newImage(scc.SiteTopUrl, abs_cur, scp.Image)
	}
}

func newImage(top string, abs_cur string, img string) string {
	if img == "" {
		return ""
	}
	if img[0] != '/' {
		img = rpath.Join(rpath.Dir(abs_cur), img)
	}

	ui, ierr := url.Parse(img)
	if ierr != nil {
		return ""
	}

	if ui.Host != "" {
		return img
	}

	u, terr := url.Parse(top)
	if terr != nil {
		return ""
	}
	if u.Host == "" {
		return ""
	}

	u.Path = rpath.Join(rpath.SetDir(u.Path), img)
	return u.String()
}

func newSiteUrl(top string, abs_cur string, site string) string {
	if site == "" {
		site = abs_cur
	} else if site[0] != '/' {
		site = rpath.Join(rpath.Dir(abs_cur), site)
	}

	pu, perr := url.Parse(site)
	if perr != nil {
		return ""
	}
	if pu.Host != "" {
		return site
	}

	u, terr := url.Parse(top)
	if terr != nil {
		return ""
	}
	if u.Host == "" {
		return ""
	}

	u.Path = rpath.Join(rpath.SetDir(u.Path), site)
	return u.String()
}

func (scc *SmCardConfig) UnmarshalTOML(decode func(interface{}) error) error {
	type rawSmCardConfig SmCardConfig
	if err := decode((*rawSmCardConfig)(scc)); err != nil {
		return err
	}

	if err := scc.validation(); err != nil {
		return err
	}

	return nil
}

var ErrInvalidSmCardParam = errors.New("invalid sm_card_config parameter")

func (scc *SmCardConfig) validation() error {
	if !scc.Enabled {
		return nil
	}

	if scc.SiteTopUrl == "" {
		return fmt.Errorf("not found site_top_url: %s", ErrInvalidSmCardParam)
	}

	{
		u, e := url.ParseRequestURI(scc.SiteTopUrl)
		if e != nil {
			return fmt.Errorf("bad site_top_url: %s", ErrInvalidSmCardParam)
		}
		if u.Host == "" {
			return fmt.Errorf("bad site_top_url: %s", ErrInvalidSmCardParam)
		}
	}

	if scc.Default.Image != "" {
		u, e := url.ParseRequestURI(scc.Default.Image)
		if e != nil {
			return fmt.Errorf("bad default_image: %s", ErrInvalidSmCardParam)
		}
		if u.Path == "" {
			return fmt.Errorf("bad default_image: %s", ErrInvalidSmCardParam)
		}
	}

	if scc.Default.SiteXcomId != "" && (len(scc.Default.SiteXcomId) < 2 || scc.Default.SiteXcomId[0] != '@') {
		return fmt.Errorf("bad default_site_xcom_id: %s", ErrInvalidSmCardParam)
	}

	return nil
}
