package group

type Group []string

func (g Group) Get(n int) string {
	if n >= len(g) || n < 0 {
		return ""
	}

	return g[n]
}

type Groups [][]string

func (g Groups) Get(m, n int) string {
	if m >= len(g) || m < 0 {
		return ""
	}

	return Group(g[m]).Get(n)
}
