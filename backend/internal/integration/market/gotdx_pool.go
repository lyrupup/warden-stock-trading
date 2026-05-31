//go:build gotdx

// 本文件仅在 `-tags gotdx` 构建时编译。
//
// 说明：bensema/gotdx 当前 release 要求 Go >= 1.26，而本仓库 module 锚定 Go 1.22，
// 因此默认构建不包含本文件，行情源回退到 stubProvider（见 stub_provider.go）。
// 接入真实 gotdx 步骤见 README「待办」。
package market

import (
	"sync"

	"github.com/bensema/gotdx"
)

// gotdxPool 解决 TDX 单连接非并发安全：每次请求借一个连接，用完归还；
// 连接异常时丢弃并下次重建（见 BACKEND.md §5 gotdx 连接池封装）。
type gotdxPool struct {
	mu    sync.Mutex
	idle  []*gotdx.Client
	max   int
	n     int
	newFn func() (*gotdx.Client, error)
}

func newGotdxPool(max int) *gotdxPool {
	return &gotdxPool{
		max: max,
		newFn: func() (*gotdx.Client, error) {
			hosts := gotdx.MainHostAddresses()
			cli := gotdx.New(
				gotdx.WithTCPAddress(hosts[0]),
				gotdx.WithTCPAddressPool(hosts[1:]...),
				gotdx.WithAutoSelectFastest(true), // 自动选最快节点
				gotdx.WithTimeoutSec(6),
			)
			return cli, cli.Connect()
		},
	}
}

func (p *gotdxPool) Get() (*gotdx.Client, error) {
	p.mu.Lock()
	if len(p.idle) > 0 {
		cli := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]
		p.mu.Unlock()
		return cli, nil
	}
	p.n++
	p.mu.Unlock()
	return p.newFn() // 新建连接（含 host 测速）
}

func (p *gotdxPool) Put(cli *gotdx.Client, broken bool) {
	if broken { // 连接异常则丢弃并重置计数，下次重建
		_ = cli.Disconnect()
		p.mu.Lock()
		p.n--
		p.mu.Unlock()
		return
	}
	p.mu.Lock()
	p.idle = append(p.idle, cli)
	p.mu.Unlock()
}
