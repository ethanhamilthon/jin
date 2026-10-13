package sounds

const Default = "bell"

type Sound struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type note struct{ frequency, start, length float64 }
type tone struct {
	Sound
	wave  string
	notes []note
}

var tones = []tone{
	{Sound{"bell", "Bell"}, "sine", []note{{880, 0, .45}, {1320, .12, .33}}},
	{Sound{"pluck", "Pluck"}, "triangle", []note{{660, 0, .25}, {990, .12, .3}}},
	{Sound{"chime", "Chime"}, "sine", []note{{1047, 0, .4}, {1568, .15, .5}}},
	{Sound{"ping", "Ping"}, "sine", []note{{1400, 0, .22}}},
	{Sound{"pop", "Pop"}, "sine", []note{{460, 0, .08}, {230, .04, .1}}},
	{Sound{"click", "Click"}, "triangle", []note{{1800, 0, .04}}},
	{Sound{"wood", "Wood"}, "triangle", []note{{350, 0, .1}, {700, 0, .06}}},
	{Sound{"drop", "Drop"}, "sine", []note{{1200, 0, .09}, {800, .06, .13}, {500, .13, .18}}},
	{Sound{"bloom", "Bloom"}, "sine", []note{{392, 0, .4}, {494, .12, .4}, {587, .24, .45}}},
	{Sound{"sparkle", "Sparkle"}, "sine", []note{{1319, 0, .14}, {1568, .08, .14}, {2093, .16, .25}}},
	{Sound{"arpeggio", "Arpeggio"}, "triangle", []note{{523, 0, .2}, {659, .1, .2}, {784, .2, .3}}},
	{Sound{"rise", "Rise"}, "sine", []note{{440, 0, .15}, {660, .1, .15}, {880, .2, .25}}},
	{Sound{"fall", "Fall"}, "sine", []note{{880, 0, .15}, {660, .1, .15}, {440, .2, .25}}},
	{Sound{"double", "Double"}, "sine", []note{{780, 0, .12}, {780, .2, .12}}},
	{Sound{"soft", "Soft"}, "sine", []note{{330, 0, .55}, {495, .08, .5}}},
	{Sound{"buzz", "Buzz"}, "triangle", []note{{220, 0, .3}, {165, .2, .35}}},
}

func Catalog() []Sound {
	list := make([]Sound, len(tones))
	for i, tone := range tones {
		list[i] = tone.Sound
	}
	return list
}

func Normalize(id string) string {
	for _, tone := range tones {
		if tone.ID == id {
			return id
		}
	}
	return Default
}

func find(id string) tone {
	id = Normalize(id)
	for _, tone := range tones {
		if tone.ID == id {
			return tone
		}
	}
	return tones[0]
}
