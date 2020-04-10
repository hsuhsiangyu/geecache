package singleflight

import "sync"

//  代表正在进行中，或已经结束的请求。
// call is an in-flight or completed Do call
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {   // 该请求已经存在
		g.mu.Unlock()
		c.wg.Wait()			// 如果请求正在进行中，则等待
		return c.val, c.err    // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)   // 发起请求前加锁
	g.m[key] = c         // 添加到 g.m，表明 key 已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn()     // 调用 fn，发起请求
	c.wg.Done()            // 请求结束

	g.mu.Lock()
	delete(g.m, key)      // 更新 g.m
	g.mu.Unlock()

	return c.val, c.err
}
