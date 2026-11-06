package crawler

func (c *Crawler) markVisited(rawURL string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.visited[rawURL]; exists {
		return false
	}

	c.visited[rawURL] = struct{}{}

	return true
}
