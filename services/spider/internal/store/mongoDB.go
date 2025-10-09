package store

import "spider/internal/entity"

type DBclient struct{}

func Page(page entity.Page) (ok bool, err error) {
	// TODO: Store HTML content in db to process later
	return
}

func NewHost() (string, bool) {
	// TODO: retrive new url from cache
	return "spaceappschallenge.org", true
}

func HostState(host entity.Host) (ok bool, err error) {
	// impl
	return
}
