package zk

import (
	"google.golang.org/grpc/resolver"
	"sync"
)

type zkResolver struct {
	scheme      string
	zkServers   []string
	zkWatchPath string
	watcher     *Watcher
	cc          resolver.ClientConn
	wg          sync.WaitGroup
}

func (r *zkResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	r.cc = cc
	var err error
	r.watcher, err = newWatcher(r.zkServers, r.zkWatchPath)
	if err != nil {
		return nil, err
	}
	r.start()
	return r, nil
}

func (r *zkResolver) Scheme() string {
	return r.scheme
}

func (r *zkResolver) start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		out := r.watcher.Watch()
		for addr := range out {
			r.cc.UpdateState(resolver.State{Addresses: addr})
		}
	}()
}

func (r *zkResolver) ResolveNow(o resolver.ResolveNowOption) {
}

func (r *zkResolver) Close() {
	r.watcher.Close()
	r.wg.Wait()
}

func RegisterResolver(scheme string, zkServers []string, srvName, srvVersion string) {
	resolver.Register(&zkResolver{
		scheme:      scheme,
		zkServers:   zkServers,
		zkWatchPath: RegistryDir + "/" + srvName + "/" + srvVersion,
	})
}
