package index

func CreateDoc(url Url, tokens []string) *Doc {
	d := Doc{
		Url: url,
		TF:  make(map[Term]uint64, defaultSize),
	}
	for _, t := range tokens {
		d.TF[t]++
	}
	return &d
}
