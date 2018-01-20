package replication

import "key-value/instance/storages"

type Target interface {
	Upsert(key string, upserter func(exist bool, prevValue string, prevVersion int64))
}

type Replicator interface {
	GetDataHandler() storages.DataHandler
}

type replicator struct {

}

func NewReplicator() Replicator {
	return &replicator{}
}

func (r *replicator) HandleRemoved(key string, version int64) {

}

func (r *replicator) HandleUpdated(key string, value string, version int64) {

}

func (r *replicator) GetDataHandler() storages.DataHandler {
	return r
}
