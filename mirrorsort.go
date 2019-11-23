package main

import (
	"sort"
)

type MirrorInfo struct {
	URL string
	PingMS int64
}
type ByPing []MirrorInfo

func (a ByPing) Len() int           { return len(a) }
func (a ByPing) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPing) Less(i, j int) bool { return a[i].PingMS < a[j].PingMS }

func PingHTTPWrapper(item interface{}) interface{} {
	url := item.(string)
	return MirrorInfo{URL: url, PingMS: PingHTTP(url)}
}

func MirrorSort(urls []string) (result []string){
	var (
		repackURL []interface{}
		repackMirror []interface{}
		mirrors []MirrorInfo
	)
	for _, url := range urls {
		repackURL = append(repackURL, url)
	}
	wp := NewWorkerPool(128, PingHTTPWrapper)
	repackMirror = wp.ProcessItems(repackURL)

	for _, mirror := range repackMirror {
		mirrors = append(mirrors, mirror.(MirrorInfo))
	}

	sort.Sort(ByPing(mirrors))

	for _, mirror := range mirrors {
		result = append(result, mirror.URL)
	}
	return
}
