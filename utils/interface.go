package utils

import "github.com/packethost/ironlib/model"

type Collector interface {
	Components() ([]*model.Component, error)
}
