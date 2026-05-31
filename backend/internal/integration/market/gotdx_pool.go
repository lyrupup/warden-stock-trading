//go:build gotdx

// 本文件仅在 `-tags gotdx` 构建时编译。
//
// gotdx 的 *gotdx.Client 单连接非并发安全（一个 TCP 连接同一时刻只能跑一次
// 请求-响应）。这里用最简连接池：借出一个连接独占使用，用完归还；连接异常时
// 丢弃并下次重建（见 BACKEND.md §5 gotdx 连接池封装）。
package market

import (
	"sync"

	"github.com/bensema/gotdx"
)

type gotdxPool struct {
	mu   sync.Mutex
	idle []*gotdx.Client
	max  int // 池中最多缓存的空闲连接数
}

func newGotdxPool(max int) *gotdxPool {
	if max <= 0 {
		max = 4
	}
	return &gotdxPool{max: max}
}

// newClient 建立并握手一个通达信主行情连接。
// 默认 Options 已内置主站地址池（gotdx.MainHostAddresses），开启测速优选最快节点。
func (p *gotdxPool) newClient() (*gotdx.Client, error) {
	cli := gotdx.New(
		gotdx.WithAutoSelectFastest(true), // 连接前对地址池做 TCP 测速，优先低延迟节点
		gotdx.WithTimeoutSec(6),
	)
	if _, err := cli.Connect(); err != nil { // Connect 返回 (*Hello1Reply, error)
		return nil, err
	}
	return cli, nil
}

// Get 借出一个可用连接：优先复用空闲连接，否则新建。
func (p *gotdxPool) Get() (*gotdx.Client, error) {
	p.mu.Lock()
	if n := len(p.idle); n > 0 {
		cli := p.idle[n-1]
		p.idle = p.idle[:n-1]
		p.mu.Unlock()
		return cli, nil
	}
	p.mu.Unlock()
	return p.newClient()
}

// Put 归还连接。broken=true 表示本次请求出错、连接可能已损坏，直接断开丢弃。
func (p *gotdxPool) Put(cli *gotdx.Client, broken bool) {
	if cli == nil {
		return
	}
	if broken {
		_ = cli.Disconnect()
		return
	}
	p.mu.Lock()
	if len(p.idle) >= p.max {
		p.mu.Unlock()
		_ = cli.Disconnect()
		return
	}
	p.idle = append(p.idle, cli)
	p.mu.Unlock()
}
