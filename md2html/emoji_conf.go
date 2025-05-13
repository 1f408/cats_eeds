package md2html

import (
	_ "embed"

	"github.com/naoina/toml"
)

//go:embed "emoji_mapping.conf"
var defaultEmojiMapping []byte

type EmojiConfig struct {
	Emoji   string
	Aliases []string
}

type EmojiMapping map[string]*EmojiConfig

func (em *EmojiMapping) Initialize() {
	type Raw EmojiMapping
	if err := toml.Unmarshal(defaultEmojiMapping, (*Raw)(em)); err != nil {
		panic("bad default emoji_mapping config file: " + err.Error())
	}
}

func (_ *EmojiMapping) MakeNew() *EmojiMapping {
	em := &EmojiMapping{}
	em.Initialize()

	return em
}

func (em *EmojiMapping) UnmarshalTOML(decode func(interface{}) error) error {
	if len(*em) == 0 {
		em.Initialize()
	}

	type Raw EmojiMapping
	return decode((*Raw)(em))
}
